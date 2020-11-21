// Code generated by "stringer -type=Actions"; DO NOT EDIT.

package chkb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ActionForward-0]
	_ = x[ActionDown-1]
	_ = x[ActionUp-2]
	_ = x[ActionTap-3]
	_ = x[ActionDoubleTap-4]
	_ = x[ActionHold-5]
	_ = x[ActionPush-6]
	_ = x[ActionPop-7]
}

const _Actions_name = "ActionForwardActionDownActionUpActionTapActionDoubleTapActionHoldActionPushActionPop"

var _Actions_index = [...]uint8{0, 13, 23, 31, 40, 55, 65, 75, 84}

func (i Actions) String() string {
	if i < 0 || i >= Actions(len(_Actions_index)-1) {
		return "Actions(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Actions_name[_Actions_index[i]:_Actions_index[i+1]]
}