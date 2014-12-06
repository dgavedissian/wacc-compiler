package frontend

import "bytes"

const STDLIB = `
begin
  bool _wacc_printBool(bool b) is
  	if b then
  	  print "true"
  	else
  	  print "false"
  	fi ;
  	return b
  end

  exit 0
end
`

func AddWaccStandardLibrary(stmt *ProgStmt) *ProgStmt {
	stdlib, asterr := GenerateAST(bytes.NewBufferString(STDLIB))
	if asterr {
		panic("Unexpected error generating stdlib!")
	}
	stmt.Funcs = append(stmt.Funcs, stdlib.Funcs...)
	return stmt
}
