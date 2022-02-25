# Cobhan FFI

Cobhan FFI is a system for enabling shared code to be written in Rust or Go and consumed from all major languages/platforms in a safe and effective way, using easy helper functions to manage any unsafe data marshaling.

## Types

* Supported types
    * int32 - 32bit signed integer
    * int64 - 64bit signed integer
    * float64 - double precision 64bit IEEE 754 floating point
    * Cobhan buffer - length delimited 8bit buffer (no null delimiters)
        * utf-8 encoded string
        * JSON
        * binary data 
* Cobhan buffer details
    * Callers provide the output buffer allocation and capacity
    * Called functions can transparently return larger values via temporary files
    * **Modern [tmpfs](https://en.wikipedia.org/wiki/Tmpfs) is entirely memory backed**
* Return values
    * Functions that return scalar values can return the value directly
        * Functions *can* use special case and return maximum positive or maximum negative or zero values to
            represent error or overflow conditions
        * Functions *can* allow scalar values to wrap
        * Functions should document their overflow / underflow behavior
