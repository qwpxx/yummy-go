package mir

import "github.com/google/uuid"

type Slot struct {
	Uuid  string
	Index uint
}

func NewSlot(index uint) Slot {
	return Slot{
		Uuid:  uuid.NewString(),
		Index: index,
	}
}

type SlotAllocator struct {
	freeSlots map[Slot]struct{}
	transform uint
}

func NewSlotAllocator() SlotAllocator {
	return SlotAllocator{
		freeSlots: make(map[Slot]struct{}),
		transform: 0,
	}
}

func (s *SlotAllocator) AllocN(n uint) []Slot {
	slots := make([]Slot, n)
	for slot := range s.freeSlots {
		slots = append(slots, slot)
		if uint(len(slots)) == n {
			break
		}
	}
	for _, slot := range slots {
		delete(s.freeSlots, slot)
	}
	for uint(len(slots)) < n {
		slots = append(slots, NewSlot(s.transform))
		s.transform += 1
	}
	return slots
}
