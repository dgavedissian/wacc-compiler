#!/bin/sh

# This script adds fields to the generated lexer.go file

#sed -i 's%  // \[NEX_END_OF_LEXER_STRUCT\]%\n  rdr *CountingReader\n // [NEX_END_OF_LEXER_STRUCT]%' $1
