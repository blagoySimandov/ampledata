#!/bin/bash
OUTPUT_FILE="workflow_analysis_data.json"
echo "[" > $OUTPUT_FILE
FIRST=1

fetch_and_append() {
    local wid="$1"
    local pid="$2"
    echo "Fetching $wid" >&2
    local history
    history=$(temporal workflow show --workflow-id "$wid" --output json)
    
    if [ $FIRST -eq 1 ]; then
        FIRST=0
    else
        echo "," >> $OUTPUT_FILE
    fi
    
    if [ -n "$pid" ]; then
        # Use jq to safely construct the JSON object to avoid escaping issues
        jq -n --arg wid "$wid" --arg pid "$pid" --argjson hist "$history" \
            '{workflow_id: $wid, parent_id: $pid, history: $hist}' >> $OUTPUT_FILE
    else
        jq -n --arg wid "$wid" --argjson hist "$history" \
            '{workflow_id: $wid, history: $hist}' >> $OUTPUT_FILE
    fi
    
    # Extract child workflow IDs
    local children
    children=$(echo "$history" | jq -r '.. | .childWorkflowExecutionStartedEventAttributes?.workflowExecution?.workflowId? | select(. != null)' | sort -u)
    
    for c in $children; do
        fetch_and_append "$c" "$wid"
    done
}

fetch_and_append "job-2950b61a-9e0d-462b-b893-9d25d44e8d50.csv" ""
fetch_and_append "job-cda094c6-e1bc-4ce9-87cd-f67a7a7c2381.csv" ""

echo "]" >> $OUTPUT_FILE
echo "Done! Output saved to $OUTPUT_FILE" >&2
