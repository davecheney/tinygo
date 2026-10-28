//go:build avr && attiny85

package avr

// The attiny85 have only TIMSK when other AVR boards have TIMSK0 and TIMSK1, etc.
// Create an alias of TIMSK, TIMSK_OCIE0A, TIMSK_OCIE0B and TIMSK_TOIE0 so we can use
// common code to mask interrupts.

var TIMSK0 = TIMSK

const (
	TIMSK0_OCIE0A = TIMSK_OCIE0A
	TIMSK0_OCIE0B = TIMSK_OCIE0B
	TIMSK0_TOIE0  = TIMSK_TOIE0
)

// The attiny85 has only TIFR, create an alias of TIFR0 for the same reason.

var TIFR0 = TIFR

const TIFR0_TOV0 = TIFR_TOV0
