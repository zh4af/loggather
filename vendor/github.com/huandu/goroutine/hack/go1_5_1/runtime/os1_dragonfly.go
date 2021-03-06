// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// From DragonFly's <sys/sysctl.h>
const (
	_CTL_HW  = 6
	_HW_NCPU = 3
)

type sigactiont struct {
	sa_sigaction uintptr
	sa_flags     int32
	sa_mask      sigset
}
