// Copyright 2015 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpuid provides access to inforamtion available with
// CPUID instruction
// All information is gathered during package initialization phase
// so package's public interface doesn't call CPUID intstruction

package cpuid

// VendorIdentificationString like "GenuineIntel" or "AuthenticAMD"
var VendorIdentificatorString string

// ProcessorBrandString like "Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz"
var ProcessorBrandString string

// SteppingId is Processor Stepping ID as described in
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var SteppingId uint32

// ProcessorType obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var ProcessorType uint32

// DisplayFamily is Family of processors obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var DisplayFamily uint32

// Display Model is Model of processor obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var DisplayModel uint32

// Cache line size in bytes
var CacheLineSize uint32

// Maximum number of addressable IDs for logical processors in this physical package
var MaxLogocalCPUId uint32

// Initial APIC ID
var InitialAPICId uint32

// Cache descriptor's array
// You can iterate like there:
// 	for _, cacheDescription := range cpuid.CacheDescriptors {
//		fmt.Printf("CacheDescriptor: %v\n", cacheDescription)
//	}
// See CacheDescriptor type for more information
var CacheDescriptors []CacheDescriptor

// Smallest monitor-line size in bytes (default is processor's monitor granularity)
var MonLineSizeMin uint32

// Largest monitor-line size in bytes (default is processor's monitor granularity)
var MonLineSizeMax uint32

// Enumeration of Monitor-Mwait extensions availability status
var MonitorEMX bool

// Supports treating interrupts as break-event for MWAIT flag
var MonitorIBE bool

// EnabledAVX flag allows to check if feature AVX is enabled by OS/BIOS
var EnabledAVX bool = false

// EnabledAVX512 flag allows to check if features AVX512xxx are enabled by OS/BIOS
var EnabledAVX512 bool = false

type CacheDescriptor struct {
	Level      int    // Cache level
	CacheType  int    // Cache type
	CacheName  string // Name
	CacheSize  int    // in KBytes (of page size for TLB)
	Ways       int    // Associativity, 0 undefined, 0xFF fully associate
	LineSize   int    // Cache line size in bytes
	Entries    int    // number of entries for TLB
	Partioning int    // partitioning
}

// ThermalSensorInterruptThresholds is the number of interrupt thresholds in digital thermal sensor.
var ThermalSensorInterruptThresholds uint32

// HasFeature to check if features from FeatureNames map are available on the current processor
func HasFeature(feature uint64) bool {
	return (featureFlags & feature) != 0
}

// HasExtendedFeature to check if features from ExtendedFeatureNames map are available on the current processor
func HasExtendedFeature(feature uint64) bool {
	return (extendedFeatureFlags & feature) != 0
}

// HasExtraFeature to check if features from ExtraFeatureNames map are available on the current processor
func HasExtraFeature(feature uint64) bool {
	return (extraFeatureFlags & feature) != 0
}

// HasThermalAndPowerFeature to check if features from ThermalAndPowerFeatureNames map are available on the current processor
func HasThermalAndPowerFeature(feature uint32) bool {
	return (thermalAndPowerFeatureFlags & feature) != 0
}

var FeatureNames = map[uint64]string{
	SSE3:         "SSE3",
	PCLMULQDQ:    "PCLMULQDQ",
	DTES64:       "DTES64",
	MONITOR:      "MONITOR",
	DSI_CPL:      "DSI_CPL",
	VMX:          "VMX",
	SMX:          "SMX",
	EST:          "EST",
	TM2:          "TM2",
	SSSE3:        "SSSE3",
	CNXT_ID:      "CNXT_ID",
	SDBG:         "SDBG",
	FMA:          "FMA",
	CX16:         "CX16",
	XTPR:         "XTPR",
	PDCM:         "PDCM",
	PCID:         "PCID",
	DCA:          "DCA",
	SSE4_1:       "SSE4_1",
	SSE4_2:       "SSE4_2",
	X2APIC:       "X2APIC",
	MOVBE:        "MOVBE",
	POPCNT:       "POPCNT",
	TSC_DEADLINE: "TSC_DEADLINE",
	AES:          "AES",
	XSAVE:        "XSAVE",
	OSXSAVE:      "OSXSAVE",
	AVX:          "AVX",
	F16C:         "F16C",
	RDRND:        "RDRND",
	HYPERVISOR:   "HYPERVISOR",
	FPU:          "FPU",
	VME:          "VME",
	DE:           "DE",
	PSE:          "PSE",
	TSC:          "TSC",
	MSR:          "MSR",
	PAE:          "PAE",
	MCE:          "MCE",
	CX8:          "CX8",
	APIC:         "APIC",
	SEP:          "SEP",
	MTRR:         "MTRR",
	PGE:          "PGE",
	MCA:          "MCA",
	CMOV:         "CMOV",
	PAT:          "PAT",
	PSE_36:       "PSE_36",
	PSN:          "PSN",
	CLFSH:        "CLFSH",
	DS:           "DS",
	ACPI:         "ACPI",
	MMX:          "MMX",
	FXSR:         "FXSR",
	SSE:          "SSE",
	SSE2:         "SSE2",
	SS:           "SS",
	HTT:          "HTT",
	TM:           "TM",
	IA64:         "IA64",
	PBE:          "PBE",
}

var ThermalAndPowerFeatureNames = map[uint32]string{ // From leaf06
	ARAT:                      "ARAT",
	PLN:                       "PLN",
	ECMD:                      "ECMD",
	PTM:                       "PTM",
	HDC:                       "HDC",
	HCFC:                      "HCFC",
	HWP:                       "HWP",
	HWP_NOTIF:                 "HWP_NOTIF",
	HWP_ACTIVITY_WINDOW:       "HWP_ACTIVITY_WINDOW",
	HWP_ENERGY_PERFORMANCE:    "HWP_ENERGY_PERFORMANCE",
	HWP_PACKAGE_LEVEL_REQUEST: "HWP_PACKAGE_LEVEL_REQUEST",
	PERFORMANCE_ENERGY_BIAS:   "PERFORMANCE_ENERGY_BIAS",
	TEMPERATURE_SENSOR:        "TEMPERATURE_SENSOR",
	TURBO_BOOST:               "TURBO_BOOST",
	TURBO_BOOST_MAX:           "TURBO_BOOST_MAX",
}

var ExtendedFeatureNames = map[uint64]string{ // From leaf07
	FSGSBASE:        "FSGSBASE",
	IA32_TSC_ADJUST: "IA32_TSC_ADJUST",
	BMI1:            "BMI1",
	HLE:             "HLE",
	AVX2:            "AVX2",
	SMEP:            "SMEP",
	BMI2:            "BMI2",
	ERMS:            "ERMS",
	INVPCID:         "INVPCID",
	RTM:             "RTM",
	PQM:             "PQM",
	DFPUCDS:         "DFPUCDS",
	MPX:             "MPX",
	PQE:             "PQE",
	AVX512F:         "AVX512F",
	AVX512DQ:        "AVX512DQ",
	RDSEED:          "RDSEED",
	ADX:             "ADX",
	SMAP:            "SMAP",
	AVX512IFMA:      "AVX512IFMA",
	PCOMMIT:         "PCOMMIT",
	CLFLUSHOPT:      "CLFLUSHOPT",
	CLWB:            "CLWB",
	INTEL_PROCESSOR_TRACE: "INTEL_PROCESSOR_TRACE",
	AVX512PF:              "AVX512PF",
	AVX512ER:              "AVX512ER",
	AVX512CD:              "AVX512CD",
	SHA:                   "SHA",
	AVX512BW:              "AVX512BW",
	AVX512VL:              "AVX512VL",
	PREFETCHWT1:           "PREFETCHWT1",
	AVX512VBMI:            "AVX512VBMI",
}

var ExtraFeatureNames = map[uint64]string{ // From leaf 8000 0001
	LAHF_LM:      "LAHF_LM",
	CMP_LEGACY:   "CMP_LEGACY",
	SVM:          "SVM",
	EXTAPIC:      "EXTAPIC",
	CR8_LEGACY:   "CR8_LEGACY",
	ABM:          "ABM",
	SSE4A:        "SSE4A",
	MISALIGNSSE:  "MISALIGNSSE",
	PREFETCHW:    "PREFETCHW",
	OSVW:         "OSVW",
	IBS:          "IBS",
	XOP:          "XOP",
	SKINIT:       "SKINIT",
	WDT:          "WDT",
	LWP:          "LWP",
	FMA4:         "FMA4",
	TCE:          "TCE",
	NODEID_MSR:   "NODEID_MSR",
	TBM:          "TBM",
	TOPOEXT:      "TOPOEXT",
	PERFCTR_CORE: "PERFCTR_CORE",
	PERFCTR_NB:   "PERFCTR_NB",
	SPM:          "SPM",
	DBX:          "DBX",
	PERFTSC:      "PERFTSC",
	PCX_L2I:      "PCX_L2I",
	FPU_2:        "FPU",
	VME_2:        "VME",
	DE_2:         "DE",
	PSE_2:        "PSE",
	TSC_2:        "TSC",
	MSR_2:        "MSR",
	PAE_2:        "PAE",
	MCE_2:        "MCE",
	CX8_2:        "CX8",
	APIC_2:       "APIC",
	SYSCALL:      "SYSCALL",
	MTRR_2:       "MTRR",
	PGE_2:        "PGE",
	MCA_2:        "MCA",
	CMOV_2:       "CMOV",
	PAT_2:        "PAT",
	PSE36:        "PSE36",
	MP:           "MP",
	NX:           "NX",
	MMXEXT:       "MMXEXT",
	MMX_2:        "MMX",
	FXSR_2:       "FXSR",
	FXSR_OPT:     "FXSR_OPT",
	PDPE1GB:      "PDPE1GB",
	RDTSCP:       "RDTSCP",
	LM:           "LM",
	_3DNOWEXT:    "3DNOWEXT",
	_3DNOW:       "3DNOW",
}

var brandStrings = map[string]int{
	"AMDisbetter!": AMD,
	"AuthenticAMD": AMD,
	"CentaurHauls": CENTAUR,
	"CyrixInstead": CYRIX,
	"GenuineIntel": INTEL,
	"TransmetaCPU": TRANSMETA,
	"GenuineTMx86": TRANSMETA,
	"Geode by NSC": NATIONALSEMICONDUCTOR,
	"NexGenDriven": NEXGEN,
	"RiseRiseRise": RISE,
	"SiS SiS SiS ": SIS,
	"UMC UMC UMC ": UMC,
	"VIA VIA VIA ": VIA,
	"Vortex86 SoC": VORTEX,
	"KVMKVMKVM":    KVM,
	"Microsoft Hv": HYPERV,
	"VMwareVMware": VMWARE,
	"XenVMMXenVMM": XEN,
}

var maxInputValue uint32
var maxExtendedInputValue uint32
var extendedModelId uint32
var extendedFamilyId uint32
var brandIndex uint32
var brandId int
var featureFlags uint64
var thermalAndPowerFeatureFlags uint32
var extendedFeatureFlags uint64
var extraFeatureFlags uint64

func cpuid_low(arg1, arg2 uint32) (eax, ebx, ecx, edx uint32) // implemented in cpuidlow_amd64.s
func xgetbv_low(arg1 uint32) (eax, edx uint32)                // implemented in cpuidlow_amd64.s
func init() {
	detectFeatures()
}

const (
	UKNOWN = iota
	AMD
	CENTAUR
	CYRIX
	INTEL
	TRANSMETA
	NATIONALSEMICONDUCTOR
	NEXGEN
	RISE
	SIS
	UMC
	VIA
	VORTEX
	KVM
	HYPERV
	VMWARE
	XEN
)

func detectFeatures() {
	leaf0()
	leaf1()
	leaf2()
	leaf3()
	leaf4()
	leaf5()
	leaf6()
	leaf7()
	leaf0x80000000()
	leaf0x80000001()
	leaf0x80000004()
	leaf0x80000005()
	leaf0x80000006()

	if HasFeature(OSXSAVE) {
		eax, _ := xgetbv_low(0)
		if (eax & 0x6) == 0x6 {
			EnabledAVX = true
		}
		if (eax & 0xE0) == 0xE0 {
			EnabledAVX512 = true
		}
	}
}

const (
	SSE3 = uint64(1) << iota
	PCLMULQDQ
	DTES64
	MONITOR
	DSI_CPL
	VMX
	SMX
	EST
	TM2
	SSSE3
	CNXT_ID
	SDBG
	FMA
	CX16
	XTPR
	PDCM
	_
	PCID
	DCA
	SSE4_1
	SSE4_2
	X2APIC
	MOVBE
	POPCNT
	TSC_DEADLINE
	AES
	XSAVE
	OSXSAVE
	AVX
	F16C
	RDRND
	HYPERVISOR
	FPU
	VME
	DE
	PSE
	TSC
	MSR
	PAE
	MCE
	CX8
	APIC
	_
	SEP
	MTRR
	PGE
	MCA
	CMOV
	PAT
	PSE_36
	PSN
	CLFSH
	_
	DS
	ACPI
	MMX
	FXSR
	SSE
	SSE2
	SS
	HTT
	TM
	IA64
	PBE
)

const (
	FSGSBASE = uint64(1) << iota
	IA32_TSC_ADJUST
	_
	BMI1
	HLE
	AVX2
	_
	SMEP
	BMI2
	ERMS
	INVPCID
	RTM
	PQM
	DFPUCDS
	MPX
	PQE
	AVX512F
	AVX512DQ
	RDSEED
	ADX
	SMAP
	AVX512IFMA
	PCOMMIT
	CLFLUSHOPT
	CLWB
	INTEL_PROCESSOR_TRACE
	AVX512PF
	AVX512ER
	AVX512CD
	SHA
	AVX512BW
	AVX512VL
	// ECX's const from there
	PREFETCHWT1
	AVX512VBMI
)

const (
	LAHF_LM = uint64(1) << iota
	CMP_LEGACY
	SVM
	EXTAPIC
	CR8_LEGACY
	ABM
	SSE4A
	MISALIGNSSE
	PREFETCHW
	OSVW
	IBS
	XOP
	SKINIT
	WDT
	_
	LWP
	FMA4
	TCE
	_
	NODEID_MSR
	_
	TBM
	TOPOEXT
	PERFCTR_CORE
	PERFCTR_NB
	SPM
	DBX
	PERFTSC
	PCX_L2I
	_
	_
	_
	// EDX features from there
	FPU_2
	VME_2
	DE_2
	PSE_2
	TSC_2
	MSR_2
	PAE_2
	MCE_2
	CX8_2
	APIC_2
	_
	SYSCALL
	MTRR_2
	PGE_2
	MCA_2
	CMOV_2
	PAT_2
	PSE36
	_
	MP
	NX
	_
	MMXEXT
	MMX_2
	FXSR_2
	FXSR_OPT
	PDPE1GB
	RDTSCP
	_
	LM
	_3DNOWEXT
	_3DNOW
)

// Thermal and Power Management features
const (
	// EAX bits 0-15
	TEMPERATURE_SENSOR        = uint32(1) << iota // Digital temperature sensor
	TURBO_BOOST                                   // Intel Turbo Boost Technology available
	ARAT                                          // APIC-Timer-always-running feature is supported if set.
	_                                             // Reserved
	PLN                                           // Power limit notification controls
	ECMD                                          // Clock modulation duty cycle extension
	PTM                                           // Package thermal management
	HWP                                           // HWP base registers (IA32_PM_ENABLE[bit 0], IA32_HWP_CAPABILITIES, IA32_HWP_REQUEST, IA32_HWP_STATUS)
	HWP_NOTIF                                     // IA32_HWP_INTERRUPT MSR
	HWP_ACTIVITY_WINDOW                           // IA32_HWP_REQUEST[bits 41:32]
	HWP_ENERGY_PERFORMANCE                        // IA32_HWP_REQUEST[bits 31:24]
	HWP_PACKAGE_LEVEL_REQUEST                     // IA32_HWP_REQUEST_PKG MSR
	_                                             // Reserved (eax bit 12)
	HDC                                           // HDC base registers IA32_PKG_HDC_CTL, IA32_PM_CTL1, IA32_THREAD_STALL MSRs
	TURBO_BOOST_MAX                               // Intel® Turbo Boost Max Technology
	_                                             // Reserved (eax bit 15)

	// ECX bits 0-15
	HCFC // Hardware Coordination Feedback Capability
	_
	_
	PERFORMANCE_ENERGY_BIAS // Processor supports performance-energy bias preference
)

const (
	NULL = iota
	DATA_CACHE
	INSTRUCTION_CACHE
	UNIFIED_CACHE
	TLB
	DTLB
	STLB
	PREFETCH
)

var leaf02Names = [...]string{
	"NULL",
	"DATA_CACHE",
	"INSTRUCTION_CACHE",
	"UNIFIED_CACHE",
	"TLB",
	"DTLB",
	"STLB",
	"PREFETCH",
}

func leaf0() {

	eax, ebx, ecx, edx := cpuid_low(0, 0)

	maxInputValue = eax

	VendorIdentificatorString = string(int32sToBytes(ebx, edx, ecx))
	brandId = brandStrings[VendorIdentificatorString]
}

func leaf1() {

	if maxInputValue < 1 {
		return
	}

	eax, ebx, ecx, edx := cpuid_low(1, 0)
	// Parse EAX
	SteppingId = (eax & 0xF)
	modelId := (eax >> 4) & 0xF
	familyId := (eax >> 8) & 0xF
	ProcessorType = (eax >> 12) & 0x3
	ExtendedModelId := (eax >> 16) & 0xF
	extendedFamilyId := (eax >> 20) & 0xFF

	DisplayFamily = familyId
	DisplayModel = modelId

	if familyId == 0xF {
		DisplayFamily = extendedFamilyId + familyId
	}

	if familyId == 0x6 || familyId == 0xF {
		DisplayModel = ExtendedModelId<<4 + modelId
	}

	// Parse EBX
	brandIndex = ebx & 0xFF
	CacheLineSize = ((ebx >> 8) & 0xFF) << 3
	MaxLogocalCPUId = (ebx >> 16) & 0xFF
	InitialAPICId = (ebx >> 24)

	// Parse ECX & EDX not needed. Ask through HasFeature function
	featureFlags = (uint64(edx) << 32) | uint64(ecx)
}

func leaf2() {

	if brandId != INTEL {
		return
	}
	if maxInputValue < 2 {
		return
	}

	bytes := int32sToBytes(cpuid_low(2, 0))

	for i := 0; i < len(bytes); i++ {
		if (i%4 == 0) && (bytes[i+3]&(1<<7) != 0) {
			i += 4
			continue
		}
		if bytes[i] == 0xFF { // it means that we should use leaf 4 for cache info
			CacheDescriptors = CacheDescriptors[0:0]
			break
		}
		CacheDescriptors = append(CacheDescriptors, leaf02Descriptors[int16(bytes[i])])
	}
}

func leaf3() {
	if brandId != INTEL {
		return
	}

	if maxInputValue < 3 {
		return
	}
	// TODO SerialNumber for < Pentium 4
}

func leaf4() {

	if brandId != INTEL {
		return
	}

	if maxInputValue < 4 {
		return
	}

	cacheId := 0
	for {
		eax, ebx, ecx, _ := cpuid_low(4, uint32(cacheId))
		cacheId++
		cacheType := eax & 0xF

		if cacheType == NULL {
			break
		}

		cacheLevel := (eax >> 5) & 0x7
		//		selfInitializingCacheLevel := eax & (1<<8)
		//		fullyAssociativeCache      := eax & (1<<9)
		//		maxNumLogicalCoresSharing  := (eax >> 14) & 0x3FF
		//		maxNumPhisCores            := (eax >> 26) & 0x3F
		systemCoherencyLineSize := (ebx & 0xFFF) + 1
		physicalLinePartions := (ebx>>12)&0x3FF + 1
		waysOfAssiociativity := (ebx>>22)&0x3FF + 1
		numberOfSets := ecx + 1
		//		writeBackInvalidate        := edx & 1
		//		cacheInclusiveness         := edx & (1<<1)
		//		complexCacheIndexing       := edx & (1<<2)
		cacheSize := (waysOfAssiociativity * physicalLinePartions *
			systemCoherencyLineSize * numberOfSets) >> 10
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{int(cacheLevel),
				int(cacheType),
				"",
				int(cacheSize),
				int(waysOfAssiociativity),
				int(systemCoherencyLineSize),
				int(numberOfSets),
				int(physicalLinePartions),
			})
	}
}

func leaf5() {
	if maxInputValue < 5 {
		return
	}

	eax, ebx, ecx, _ := cpuid_low(4, 0) // TODO process EDX with C0-C7 C-states
	MonLineSizeMax = eax & (0xFFFF)
	MonLineSizeMax = ebx & (0xFFFF)
	MonitorEMX = (ecx & (1 << 0)) != 0
	MonitorIBE = (ecx & (1 << 1)) != 0

}

func leaf6() {
	// Thermal and Power Management Features for Intel
	if maxInputValue < 6 {
		return
	}

	eax, ebx, ecx, _ := cpuid_low(6, 0)
	thermalAndPowerFeatureFlags = (eax & 0xFFFF) | (ecx << 16)
	ThermalSensorInterruptThresholds = ebx & 7
}

func leaf7() {
	_, ebx, ecx, _ := cpuid_low(7, 0)
	extendedFeatureFlags = (uint64(ecx) << 32) | uint64(ebx)
}

func leaf0x80000000() {
	maxExtendedInputValue, _, _, _ = cpuid_low(0x80000000, 0)
}

func leaf0x80000001() {
	if maxExtendedInputValue < 0x80000001 {
		return
	}
	_, _, ecx, edx := cpuid_low(0x80000001, 0)
	//extendedProcessorSignatureAndFeatureBits := eax
	extraFeatureFlags = (uint64(edx) << 32) | uint64(ecx)
}

// leaf0x80000004 looks at the Processor Brand String in leaves 0x80000002 through 0x80000004
func leaf0x80000004() {
	if maxExtendedInputValue < 0x80000004 {
		return
	}

	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000002, 0)))
	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000003, 0)))
	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000004, 0)))
}

func leaf0x80000005() {
	// AMD L1 Cache and TLB Information
	if maxExtendedInputValue < 0x80000005 {
		return
	}

	if brandId != AMD {
		return
	}

	eax, ebx, ecx, edx := cpuid_low(0x80000005, 0)

	L1DTlb2and4MAssoc := (eax >> 24) & 0xFF
	L1DTlb2and4MSize := (eax >> 16) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DTLB,
			"DTLB 2M/4M",
			2 * 1024,
			int(L1DTlb2and4MAssoc),
			-1,
			int(L1DTlb2and4MSize),
			0,
		})

	L1ITlb2and4MAssoc := (eax >> 8) & 0xFF
	L1ITlb2and4MSize := (eax) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			TLB,
			"ITLB 2M/4M",
			2 * 1024,
			int(L1ITlb2and4MAssoc),
			-1,
			int(L1ITlb2and4MSize),
			0,
		})

	L1DTlb4KAssoc := (ebx >> 24) & 0xFF
	L1DTlb4KSize := (ebx >> 16) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DTLB,
			"DTLB 4K",
			4,
			int(L1DTlb4KAssoc),
			-1,
			int(L1DTlb4KSize),
			0,
		})

	L1ITlb4KAssoc := (ebx >> 8) & 0xFF
	L1ITlb4KSize := (ebx) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			TLB,
			"ITLB 4K",
			4,
			int(L1ITlb4KAssoc),
			-1,
			int(L1ITlb4KSize),
			0,
		})

	L1DcSize := (ecx >> 24) & 0xFF
	L1DcAssoc := (ecx >> 16) & 0xFF
	L1DcLinesPerTag := (ecx >> 8) & 0xFF
	L1DcLineSize := (ecx >> 0) & 0xFF
	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DATA_CACHE,
			"L1 Data cache",
			int(L1DcSize),
			int(L1DcAssoc),
			int(L1DcLineSize),
			-1,
			int(L1DcLinesPerTag),
		})

	L1IcSize := (edx >> 24) & 0xFF
	L1IcAssoc := (edx >> 16) & 0xFF
	L1IcLinesPerTag := (edx >> 8) & 0xFF
	L1IcLineSize := (edx >> 0) & 0xFF
	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			INSTRUCTION_CACHE,
			"L1 Instruction cache",
			int(L1IcSize),
			int(L1IcAssoc),
			int(L1IcLineSize),
			-1,
			int(L1IcLinesPerTag),
		})
}

func leaf0x80000006() {

	if maxExtendedInputValue < 0x80000006 {
		return
	}

	var associativityEncodings = map[uint]uint{
		0x00: 0,
		0x01: 1,
		0x02: 2,
		0x04: 4,
		0x06: 8,
		0x08: 16,
		0x0A: 32,
		0x0B: 48,
		0x0C: 64,
		0x0D: 96,
		0x0E: 128,
		0x0F: 0xFF, // - Fully associative
	}

	eax, ebx, ecx, edx := cpuid_low(0x80000006, 0)

	if brandId == INTEL {

		CacheLineSize := (ecx >> 0) & 0xFF
		L2Associativity := uint((ecx >> 12) & 0xF)
		CacheSize := (ecx >> 16) & 0xFFFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				0,
				"Cache info from leaf 0x80000006 for Intel",
				int(CacheSize),
				int(associativityEncodings[L2Associativity]),
				int(CacheLineSize),
				-1,
				0,
			})
	}

	if brandId == AMD {

		L2DTlb2and4MAssoc := uint((eax >> 28) & 0xF)
		L2DTlb2and4MSize := (eax >> 16) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DTLB,
				"DTLB 2M/4M",
				2 * 1024,
				int(associativityEncodings[L2DTlb2and4MAssoc]),
				-1,
				int(L2DTlb2and4MSize),
				0,
			})

		L2ITlb2and4MAssoc := uint((eax >> 12) & 0xF)
		L2ITlb2and4MSize := (eax) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				TLB,
				"ITLB 2M/4M",
				2 * 1024,
				int(associativityEncodings[L2ITlb2and4MAssoc]),
				-1,
				int(L2ITlb2and4MSize),
				0,
			})

		L2DTlb4KAssoc := uint((ebx >> 28) & 0xF)
		L2DTlb4KSize := (ebx >> 16) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DTLB,
				"DTLB 4K",
				4,
				int(associativityEncodings[L2DTlb4KAssoc]),
				-1,
				int(L2DTlb4KSize),
				0,
			})

		L2ITlb4KAssoc := uint((ebx >> 12) & 0xF)
		L2ITlb4KSize := (ebx) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				TLB,
				"ITLB 4K",
				4,
				int(associativityEncodings[L2ITlb4KAssoc]),
				-1,
				int(L2ITlb4KSize),
				0,
			})

		L2Size := (ecx >> 16) & 0xFFFF
		L2Assoc := uint((ecx >> 12) & 0xF)
		L2LinesPerTag := (ecx >> 8) & 0xF
		L2LineSize := (ecx >> 0) & 0xFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DATA_CACHE,
				"L2 Data cache",
				int(L2Size),
				int(associativityEncodings[L2Assoc]),
				int(L2LineSize),
				-1,
				int(L2LinesPerTag),
			})

		L3Size := ((edx >> 18) & 0xF) * 512
		L3Assoc := uint((edx >> 12) & 0xF)
		L3LinesPerTag := (edx >> 8) & 0xF
		L3LineSize := (edx >> 0) & 0xFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{3,
				DATA_CACHE,
				"L3 Data cache",
				int(L3Size),
				int(associativityEncodings[L3Assoc]),
				int(L3LineSize),
				-1,
				int(L3LinesPerTag),
			})
	}
}

// TODO split fused descritops with bits in high key's byte like for 0x49
var leaf02Descriptors = map[int16]CacheDescriptor{
	0x01: {-1, TLB, "Instruction TLB", 4, 4, -1, 32, 0},
	0x02: {-1, TLB, "Instruction TLB", 4 * 1024, 0xFF, -1, 2, 0},
	0x03: {-1, TLB, "Data TLB", 4, 4, -1, 64, 0},
	0x04: {-1, TLB, "Data TLB", 4 * 1024, 4, -1, 8, 0},
	0x05: {-1, TLB, "Data TLB1", 4 * 1024, 4, -1, 32, 0},
	0x06: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 8, 4, 32, -1, 0},
	0x08: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 16, 4, 32, -1, 0},
	0x09: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 32, 4, 64, -1, 0},
	0x0A: {1, DATA_CACHE, "1st-level data cache", 8, 2, 32, -1, 0},
	0x0B: {-1, TLB, "Instruction TLB", 4 * 1024, 4, -1, 4, 0},
	0x0C: {1, DATA_CACHE, "1st-level data cache", 16, 4, 32, -1, 0},
	0x0D: {1, DATA_CACHE, "1st-level data cache", 16, 4, 64, -1, 0},
	0x0E: {1, DATA_CACHE, "1st-level data cache", 24, 6, 64, -1, 0},
	0x1D: {2, DATA_CACHE, "2nd-level cache", 128, 2, 64, -1, 0},
	0x21: {2, DATA_CACHE, "2nd-level cache", 256, 8, 64, -1, 0},
	0x22: {3, DATA_CACHE, "3nd-level cache", 512, 4, 64, -1, 2},
	0x23: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 8, 64, -1, 2},
	0x24: {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 16, 64, -1, 0},
	0x25: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 8, 64, -1, 2},
	0x29: {3, DATA_CACHE, "2nd-level cache", 4 * 1024, 8, 64, -1, 2},
	0x2C: {1, DATA_CACHE, "1st-level cache", 32, 8, 64, -1, 0},
	0x30: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 32, 8, 64, -1, 0},
	0x40: {-1, DATA_CACHE, "No 2nd-level cache or, if processor contains a " +
		"valid 2nd-level cache, no 3rd-level cache", -1, -1, -1, -1, 0},
	0x41: {2, DATA_CACHE, "2nd-level cache", 128, 4, 32, -1, 0},
	0x42: {2, DATA_CACHE, "2nd-level cache", 256, 4, 32, -1, 0},
	0x43: {2, DATA_CACHE, "2nd-level cache", 512, 4, 32, -1, 0},
	0x44: {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 4, 32, -1, 0},
	0x45: {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 4, 32, -1, 0},
	0x46: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 4, 64, -1, 0},
	0x47: {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 8, 64, -1, 0},
	0x48: {2, DATA_CACHE, "2nd-level cache", 3 * 1024, 12, 64, -1, 0},
	0x49: {2, DATA_CACHE, "2nd-level cache", 4 * 1024, 16, 64, -1, 0},
	// (Intel Xeon processor MP, Family 0FH, Model 06H)
	(0x49 | (1 << 8)): {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 16, 64, -1, 0},
	0x4A:              {3, DATA_CACHE, "3nd-level cache", 6 * 1024, 12, 64, -1, 0},
	0x4B:              {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 16, 64, -1, 0},
	0x4C:              {3, DATA_CACHE, "3nd-level cache", 12 * 1024, 12, 64, -1, 0},
	0x4D:              {3, DATA_CACHE, "3nd-level cache", 16 * 1024, 16, 64, -1, 0},
	0x4E:              {2, DATA_CACHE, "3nd-level cache", 6 * 1024, 24, 64, -1, 0},
	0x4F:              {-1, TLB, "Instruction TLB", 4, -1, -1, 32, 0},
	0x50:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 64, 0},
	0x51:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 128, 0},
	0x52:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 256, 0},
	0x55:              {-1, TLB, "Instruction TLB: 2-MByte or 4-MByte pages", 2 * 1024, 0xFF, -1, 7, 0},
	0x56:              {-1, TLB, "Data TLB0", 4 * 1024, 4, -1, 16, 0},
	0x57:              {-1, TLB, "Data TLB0", 4, 4, -1, 16, 0},
	0x59:              {-1, TLB, "Data TLB0", 4, 0xFF, -1, 16, 0},
	0x5A:              {-1, TLB, "Data TLB0 2-MByte or 4 MByte pages", 2 * 1024, 4, -1, 32, 0},
	0x5B:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 64, 0},
	0x5C:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 128, 0},
	0x5D:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 256, 0},
	0x60:              {1, DATA_CACHE, "1st-level data cache", 16, 8, 64, -1, 0},
	0x61:              {-1, TLB, "Instruction TLB", 4, 0xFF, -1, 48, 0},
	0x63:              {-1, TLB, "Data TLB", 1 * 1024 * 1024, 4, -1, 4, 0},
	0x66:              {1, DATA_CACHE, "1st-level data cache", 8, 4, 64, -1, 0},
	0x67:              {1, DATA_CACHE, "1st-level data cache", 16, 4, 64, -1, 0},
	0x68:              {1, DATA_CACHE, "1st-level data cache", 32, 4, 64, -1, 0},
	0x70:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 12, 8, -1, -1, 0},
	0x71:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 16, 8, -1, -1, 0},
	0x72:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 32, 8, -1, -1, 0},
	0x76:              {-1, TLB, "Instruction TLB: 2M/4M pages", 2 * 1024, 0xFF, -1, 8, 0},
	0x78:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 4, 64, -1, 0},
	0x79:              {2, DATA_CACHE, "2nd-level cache", 128, 8, 64, -1, 2},
	0x7A:              {2, DATA_CACHE, "2nd-level cache", 256, 8, 64, -1, 2},
	0x7B:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 64, -1, 2},
	0x7C:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 64, -1, 2},
	0x7D:              {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 8, 64, -1, 0},
	0x7F:              {2, DATA_CACHE, "2nd-level cache", 512, 2, 64, -1, 0},
	0x80:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 64, -1, 0},
	0x82:              {2, DATA_CACHE, "2nd-level cache", 256, 8, 32, -1, 0},
	0x83:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 32, -1, 0},
	0x84:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 32, -1, 0},
	0x85:              {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 8, 32, -1, 0},
	0x86:              {2, DATA_CACHE, "2nd-level cache", 512, 4, 32, -1, 0},
	0x87:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 64, -1, 0},
	0xA0:              {-1, DTLB, "DTLB", 4, 0xFF, -1, 32, 0},
	0xB0:              {-1, TLB, "Instruction TLB", 4, 4, -1, 128, 0},
	0xB1: {-1, TLB, "Instruction TLB 2M pages 4 way 8 entries or" +
		"4M pages 4-way, 4 entries", 2 * 1024, 4, -1, 8, 0},
	0xB2: {-1, TLB, "Instruction TLB", 4, 4, -1, 64, 0},
	0xB3: {-1, TLB, "Data TLB", 4, 4, -1, 128, 0},
	0xB4: {-1, TLB, "Data TLB1", 4, 4, -1, 256, 0},
	0xB5: {-1, TLB, "Instruction TLB", 4, 8, -1, 64, 0},
	0xB6: {-1, TLB, "Instruction TLB", 4, 8, -1, 128, 0},
	0xBA: {-1, TLB, "Data TLB1", 4, 4, -1, 64, 0},
	0xC0: {-1, TLB, "Data TLB: 4 KByte and 4 MByte pages", 4, 4, -1, 8, 0},
	0xC1: {-1, STLB, "Shared 2nd-Level TLB: 4Kbyte and 2Mbyte pages", 4, 8, -1, 1024, 0},
	0xC2: {-1, DTLB, "DTLB 4KByte/2 MByte pages", 4, 4, -1, 16, 0},
	0xC3: {-1, STLB, "Shared 2nd-Level TLB: " +
		"4 KByte /2 MByte pages, 6-way associative, 1536 entries." +
		"Also 1GBbyte pages, 4-way,16 entries.", 4, 6, -1, 1536, 0},
	0xCA: {-1, STLB, "Shared 2nd-Level TLB", 4, 4, -1, 512, 0},
	0xD0: {3, DATA_CACHE, "3nd-level cache", 512, 4, 64, -1, 0},
	0xD1: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 4, 64, -1, 0},
	0xD2: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 4, 64, -1, 0},
	0xD6: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 8, 64, -1, 0},
	0xD7: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 8, 64, -1, 0},
	0xD8: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 8, 64, -1, 0},
	0xDC: {3, DATA_CACHE, "3nd-level cache", 1 * 1536, 12, 64, -1, 0},
	0xDD: {3, DATA_CACHE, "3nd-level cache", 3 * 1024, 12, 64, -1, 0},
	0xDE: {3, DATA_CACHE, "3nd-level cache", 6 * 1024, 12, 64, -1, 0},
	0xE2: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 16, 64, -1, 0},
	0xE3: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 16, 64, -1, 0},
	0xE4: {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 16, 64, -1, 0},
	0xEA: {3, DATA_CACHE, "3nd-level cache", 12 * 1024, 24, 64, -1, 0},
	0xEB: {3, DATA_CACHE, "3nd-level cache", 18 * 1024, 24, 64, -1, 0},
	0xEC: {3, DATA_CACHE, "3nd-level cache", 24 * 1024, 24, 64, -1, 0},
	0xF0: {-1, PREFETCH, "", 64, -1, -1, -1, 0},
	0xF1: {-1, PREFETCH, "", 128, -1, -1, -1, 0},
	0xFF: {-1, NULL, "CPUID leaf 2 does not report cache descriptor " +
		"information, use CPUID leaf 4 to query cache parameters",
		-1, -1, -1, -1, 0},
}

func int32sToBytes(args ...uint32) []byte {
	var result []byte

	for _, arg := range args {
		result = append(result,
			byte((arg)&0xFF),
			byte((arg>>8)&0xFF),
			byte((arg>>16)&0xFF),
			byte((arg>>24)&0xFF))
	}

	return result
}
