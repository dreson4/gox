//go:build ios

package haptics

/*
extern void goxHapticImpact(int style);
extern void goxHapticNotify(int typ);
extern void goxHapticSelection(void);
*/
import "C"

func platformImpact(style ImpactStyle) {
	C.goxHapticImpact(C.int(style))
}

func platformNotify(typ NotificationType) {
	C.goxHapticNotify(C.int(typ))
}

func platformSelection() {
	C.goxHapticSelection()
}
