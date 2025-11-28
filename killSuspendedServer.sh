#!/bin/bash

count=0
# Scan for open files specifically network files on port 8080
# Put into array called items
lsof -i tcp:8080 | while IFS= read -a items; do
    for value in “${items[@]}”; do
        ((count++))
        # Get the modulo of 11 because thats where the Process ID column values are typically founded
        if [ $((count % 11)) -eq 0 ]; then
            ((count++))
            kill -9 $value
        fi
    done
done
