//go:build tinygo.riscv && virt && scheduler.cores

package runtime

import "device/riscv"

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	hartID := currentCPU()

	// Mark the current hart as sleeping.
	waitingCores |= 1 << hartID

	// If this is the last core awake and is going to sleep, the scheduler is
	// deadlocked.
	// We can do this check since this is not baremetal: there won't be any
	// external interrupts that might unblock a goroutine.
	if waitingCores == (1<<numCPU)-1 {
		runtimeFatal("all cores are sleeping - deadlock!")
	}

	// Need to disable interrupts while saving the checkpoint, otherwise if the
	// software interrupt happens earlier for another reason (e.g. a GC cycle)
	// it will see an incomplete checkpoint and the schedulerLock might not be
	// unlocked yet. That will lead to an invalid state.
	riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)
	if schedulerWaitCheckpoints[hartID].Save() {
		schedulerLock.Unlock()
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Wait until we get awoken :)
		for {
			riscv.Asm("wfi")
		}
	}

	// We got awoken again. We need to lock the scheduler again before
	// returning.
	schedulerLock.Lock()
}

// End schedulerUnlockAndWait on the given hart.
func schedulerEndWait(hart uint32) {
	schedulerWakePending[hart].Store(1)
	signalHart(hart)
}
