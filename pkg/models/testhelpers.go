package models

import (
	"testing"
)

func SingleDataCompareHelper(result []*SingleData, wanted []*SingleData, t *testing.T) {
	// Raise no errors if two []SingleData are identical, raise errors if they are not
	if len(result) != len(wanted) {
		t.Errorf("Input and output SingleData differ in length. Wanted %v, got %v", len(wanted), len(result))
		return
	}
	for udx := range wanted {
		if result[udx].Name != wanted[udx].Name {
			t.Errorf("Input and output SingleData have different Pvs. Wanted %v, got %v", wanted[udx].Name, result[udx].Name)
		}

		switch resultv := result[udx].Values.(type) {
		case *Scalars:
			wantedv := wanted[udx].Values.(*Scalars)
			if len(wantedv.Times) != len(resultv.Times) {
				t.Errorf("Input and output arrays' times differ in length. Wanted %v, got %v", len(wantedv.Times), len(resultv.Times))
				return
			}
			if len(wantedv.Values) != len(resultv.Values) {
				t.Errorf("Input and output arrays' values differ in length. Wanted %v, got %v", len(wantedv.Values), len(resultv.Values))
				return
			}
			for idx := range wantedv.Values {
				if resultv.Times[idx] != wantedv.Times[idx] {
					t.Errorf("Times at index %v do not match, Wanted %v, got %v", idx, wantedv.Times[idx], resultv.Times[idx])
				}
				if resultv.Values[idx] != wantedv.Values[idx] {
					t.Errorf("Values at index %v do not match, Wanted %v, got %v", idx, wantedv.Values[idx], resultv.Values[idx])
				}
			}
		case *Arrays:
			wantedv := wanted[udx].Values.(*Arrays)
			if len(wantedv.Times) != len(resultv.Times) {
				t.Errorf("Input and output arrays' times differ in length. Wanted %v, got %v", len(wantedv.Times), len(resultv.Times))
				return
			}
			if len(wantedv.Values) != len(resultv.Values) {
				t.Errorf("Input and output arrays' values differ in length. Wanted %v, got %v", len(wantedv.Values), len(resultv.Values))
				return
			}
			for idx := range wantedv.Values {
				if resultv.Times[idx] != wantedv.Times[idx] {
					t.Errorf("Times at index %v do not match, Wanted %v, got %v", idx, wantedv.Times[idx], resultv.Times[idx])
				}
				for idy := range wantedv.Values[idx] {
					if resultv.Values[idx][idy] != wantedv.Values[idx][idy] {
						t.Errorf("Values at index %v do not match, Wanted %v, got %v", idx, wantedv.Values[idx][idy], resultv.Values[idx][idy])
					}
				}
			}
		default:
			t.Fatalf("Response Values are invalid")
		}
	}
}
