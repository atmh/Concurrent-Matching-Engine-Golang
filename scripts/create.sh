#!/bin/bash

# Set the directory containing the target files
input_dir="customTests"

# Set the directory for the randomized output files
output_dir="randomizedTests"

# Create the output directory if it doesn't already exist
mkdir -p "$output_dir"

# Loop through each file in the input directory
for input_file in "$input_dir"/*; do 
    (
    # Get the filename without the directory path
    filename=$(basename "$input_file")
    # Determine which lines to randomize
    if [ "$(head -n 1 "$input_file" | cut -c1)" = "#" ]; then
        start_line=4
    else
        start_line=3
    fi
    end_line=$(grep -nve '^\s*$' "$input_file" | tail -n 3 | head -n 1 | cut -d: -f1)
    # Create three randomized versions of the file
    for i in {1..10}; do
        (
        # Create the output filename
        output_file="${output_dir}/${i}-${filename}"
        # Copy the initial lines up to the start line to the output file
        head -n "$((start_line-1))" "$input_file" > "$output_file"
        # Get a list of line numbers to randomize and shuffle them
        line_numbers=$(seq "$start_line" "$end_line")
        shuffled_line_numbers=$(echo "$line_numbers" | shuf)
        # Loop through the shuffled line numbers and copy the corresponding lines to the output file
        while read -r line_number; do
            sed "${line_number}q;d" "$input_file" >> "$output_file"
        done <<< "$(echo "$shuffled_line_numbers")"
        # Copy the remaining lines from the end line onwards to the output file
        tail -n "+$((end_line+1))" "$input_file" >> "$output_file"
        # Print a message indicating the output filename
        echo "Created randomized file: $output_file" 
        ) &
    done
    ) &
done
