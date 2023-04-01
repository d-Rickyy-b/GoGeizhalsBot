package bot

import "testing"

func Test_tempPriceStore_storePrice(t1 *testing.T) {
	t := tempPriceStore{}

	_, exists := t.getPrice(1234, "de")
	if exists {
		t1.Errorf("Price shouldn't exist yet")
	}

	t.storePrice(1234, "de", 1.0)
	if t.store[1234]["de"] != 1.0 {
		t1.Errorf("tempPriceStore.storePrice() failed")
	}

	_, exists2 := t.getPrice(1234, "de")
	if !exists2 {
		t1.Errorf("Price must exist now!")
	}

	_, exists3 := t.getPrice(1234, "at")
	if exists3 {
		t1.Errorf("Price shouldn't exist for other location!")
	}

	t.storePrice(1234, "at", 34.56)
	if t.store[1234]["at"] != 34.56 {
		t1.Errorf("tempPriceStore.storePrice() failed")
	}

	p, exists4 := t.getPrice(1234, "at")
	if !exists4 || p != 34.56 {
		t1.Errorf("Price shouldn't exist for other location!")
	}
}
