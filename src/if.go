package main

type Instr interface {
}

type SkipInstr struct {
	Next Instr
}

type LoadInstr struct {
	Next Instr
	Addr int
	Reg  int
}

type StoreInstr struct {
	Next Instr
}

func GenerateIntermediateForm() Instr {
	return nil
}
