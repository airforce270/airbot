// Package logs provides a common set of logger functions for the bot.
package logs

import (
	baselog "log"
	"os"
)

// Logger is a singleton logger to use for logging.
var Logger = baselog.New(os.Stderr, "", baselog.Ldate|baselog.Ltime|baselog.Lmicroseconds)

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...any) { Logger.Printf(format, v...) }

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...any) { Logger.Print(v...) }

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...any) { Logger.Println(v...) }

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func Fatal(v ...any) { Logger.Fatal(v...) }

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...any) { Logger.Fatalf(format, v...) }

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func Fatalln(v ...any) { Logger.Fatalln(v...) }

// Panic is equivalent to l.Print() followed by a call to panic().
func Panic(v ...any) { Logger.Panic(v...) }

// Panicf is equivalent to l.Printf() followed by a call to panic().
func Panicf(format string, v ...any) { Logger.Panicf(format, v...) }

// Panicln is equivalent to l.Println() followed by a call to panic().
func Panicln(v ...any) { Logger.Panicln(v...) }
