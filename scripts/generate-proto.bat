@echo off
REM Generate Go code from proto files for Windows
REM Requires: protoc, protoc-gen-go, protoc-gen-go-grpc

echo 🔧 Generating gRPC code from proto files...

REM Create output directory if it doesnt exist
if not exist "proto\example" mkdir proto\example

REM Generate Go code
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/example/example.proto proto/nuha-auth/auth.proto

if %ERRORLEVEL% EQU 0 (
    echo ✅ gRPC code generated successfully!
) else (
    echo ❌ Failed to generate gRPC code
    exit /b 1
)
