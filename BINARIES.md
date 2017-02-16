# Building "binaries"

For perl toolchain-free distribution.

Install PAR::Packer first, then:

## Mac

pp -M IO::Socket::SSL -o dau-mac dau

## Linux

pp -M IO::Socket::SSL -o dau-linux dau

## Windows

pp -M IO::Socket::SSL -o dau.exe dau
