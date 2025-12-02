package main

import (
	"fmt"
	"time"
)

type BookingState string

const (
	StateIdle             BookingState = "Idle"
	StateRoomSelected     BookingState = "RoomSelected"
	StateBookingConfirmed BookingState = "BookingConfirmed"
	StatePaid             BookingState = "Paid"
	StateBookingCancelled BookingState = "BookingCancelled"
)

type BookingEvent string

const (
	EventSelectRoom     BookingEvent = "selectRoom"
	EventConfirmBooking BookingEvent = "confirmBooking"
	EventPay            BookingEvent = "pay"
	EventCancel         BookingEvent = "cancel"
	EventChangeRoom     BookingEvent = "changeRoom"
)

type Room struct {
	ID    int
	Type  string
	Price float64
}

type Booking struct {
	ID        int
	UserID    int
	Room      *Room
	State     BookingState
	CreatedAt time.Time
	PaidAt    time.Time
	Total     float64
}

type BookingHistory struct {
	Bookings []*Booking
}

func (bh *BookingHistory) Add(b *Booking) {
	bh.Bookings = append(bh.Bookings, b)
}

var discounts = map[string]float64{
	"LOYALTY10": 10.0,
	"HOLIDAY15": 15.0,
}

type HotelBookingSystem struct {
	nextBookingID int
	history       *BookingHistory
}

func NewHotelBookingSystem() *HotelBookingSystem {
	return &HotelBookingSystem{
		nextBookingID: 1,
		history:       &BookingHistory{},
	}
}

func (h *HotelBookingSystem) canTransition(from, to BookingState, event BookingEvent) bool {
	transitions := map[BookingState]map[BookingEvent]BookingState{
		StateIdle: {
			EventSelectRoom: StateRoomSelected,
		},
		StateRoomSelected: {
			EventConfirmBooking: StateBookingConfirmed,
			EventChangeRoom:     StateRoomSelected,
			EventCancel:         StateBookingCancelled,
		},
		StateBookingConfirmed: {
			EventPay:    StatePaid,
			EventCancel: StateBookingCancelled,
		},
	}
	return transitions[from][event] == to
}

func (h *HotelBookingSystem) Transition(booking *Booking, event BookingEvent, newRoom *Room, promoCode string) error {
	var newState BookingState

	switch event {
	case EventSelectRoom:
		if booking.State != StateIdle {
			return fmt.Errorf("cannot select room from state %s", booking.State)
		}
		booking.Room = newRoom
		newState = StateRoomSelected

	case EventChangeRoom:
		if booking.State != StateRoomSelected {
			return fmt.Errorf("changing room is only available in RoomSelected state")
		}
		booking.Room = newRoom
		newState = StateRoomSelected

	case EventConfirmBooking:
		if booking.State != StateRoomSelected {
			return fmt.Errorf("confirmation is only possible after selecting a room")
		}
		newState = StateBookingConfirmed

	case EventCancel:
		if booking.State == StatePaid {
			return fmt.Errorf("cannot cancel a paid booking")
		}
		newState = StateBookingCancelled

	case EventPay:
		if booking.State != StateBookingConfirmed {
			return fmt.Errorf("payment is only possible after confirmation")
		}
		total := booking.Room.Price
		if discount, ok := discounts[promoCode]; ok {
			total *= (1 - discount/100)
			fmt.Printf("Promo code %s applied. Discount: %.0f%%\n", promoCode, discount)
		}
		booking.Total = total
		booking.PaidAt = time.Now()
		newState = StatePaid

	default:
		return fmt.Errorf("unknown event: %s", event)
	}

	if !h.canTransition(booking.State, newState, event) {
		return fmt.Errorf("invalid transition: %s -> %s", booking.State, event)
	}

	fmt.Printf("Booking #%d: %s -> %s\n", booking.ID, booking.State, newState)
	booking.State = newState

	if newState == StatePaid || newState == StateBookingCancelled {
		h.history.Add(booking)
	}

	return nil
}

func (h *HotelBookingSystem) NewBooking(userID int) *Booking {
	b := &Booking{
		ID:        h.nextBookingID,
		UserID:    userID,
		State:     StateIdle,
		CreatedAt: time.Now(),
	}
	h.nextBookingID++
	return b
}

func main() {
	system := NewHotelBookingSystem()

	standard := &Room{ID: 101, Type: "standard", Price: 5000}
	deluxe := &Room{ID: 201, Type: "deluxe", Price: 10000}

	fmt.Println("=== Scenario 1: Successful booking ===")
	booking1 := system.NewBooking(1001)
	system.Transition(booking1, EventSelectRoom, standard, "")
	system.Transition(booking1, EventConfirmBooking, nil, "")
	system.Transition(booking1, EventPay, nil, "LOYALTY10")

	fmt.Println("\n=== Scenario 2: Cancellation before payment ===")
	booking2 := system.NewBooking(1002)
	system.Transition(booking2, EventSelectRoom, deluxe, "")
	system.Transition(booking2, EventCancel, nil, "")

	fmt.Println("\n=== Scenario 3: Change room ===")
	booking3 := system.NewBooking(1003)
	system.Transition(booking3, EventSelectRoom, standard, "")
	system.Transition(booking3, EventChangeRoom, deluxe, "")
	system.Transition(booking3, EventConfirmBooking, nil, "")
	system.Transition(booking3, EventPay, nil, "")

	fmt.Println("\n=== Booking History ===")
	for _, b := range system.history.Bookings {
		status := "CANCELLED"
		if b.State == StatePaid {
			status = "PAID"
		}
		fmt.Printf("ID: %d | Room: %d | Total: %.0f | Status: %s\n",
			b.ID, b.Room.ID, b.Total, status)
	}
}
