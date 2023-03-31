#!/bin/bash

i=1

for file in customTests/*
do
    (
        output=$(../grader ../engine < "$file" 2>&1 | tail -1 | tr -d '\0')
       if [ "$output" == "test passed." ]
        then
            echo "Test case $file: ✓"
        else
            echo $file : $output
        fi
    ) &
    
    i=$((i+1))
done

wait

echo "done with custom tests"