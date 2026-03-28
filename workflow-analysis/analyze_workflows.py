import json
import os
import argparse
import sys
from google import genai
from google.genai import types


def load_workflows(file_path):
    with open(file_path, "r") as f:
        return json.load(f)


def extract_meaningful_events(history_json):
    """
    Reduces the payload size by keeping only essential events like starts, completions, and failures.
    """
    events = history_json.get("events", [])
    meaningful_events = []

    for event in events:
        event_type = event.get("eventType")
        if event_type in [
            "EVENT_TYPE_WORKFLOW_EXECUTION_STARTED",
            "EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED",
            "EVENT_TYPE_WORKFLOW_EXECUTION_FAILED",
            "EVENT_TYPE_ACTIVITY_TASK_SCHEDULED",
            "EVENT_TYPE_ACTIVITY_TASK_STARTED",
            "EVENT_TYPE_ACTIVITY_TASK_COMPLETED",
            "EVENT_TYPE_ACTIVITY_TASK_FAILED",
            "EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED",
            "EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED",
            "EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED",
        ]:
            meaningful_events.append(event)

    return {"events": meaningful_events}


def analyze_child(client, child_data):
    print(f"Analyzing child workflow: {child_data['workflow_id']}...")

    # Trim history to save tokens and reduce noise
    trimmed_history = extract_meaningful_events(child_data["history"])

    prompt = f"""
    You are an expert Temporal workflow analyst. Analyze the following child workflow history.
    
    Workflow ID: {child_data["workflow_id"]}
    
    Identify:
    1. The core task/activity it was trying to perform.
    2. Key inputs and outputs (summarize the payloads).
    3. Any activity or workflow failures, timeouts, and retries.
    4. Provide a brief assessment of its execution (e.g., successful, slow, failed).
    
    Workflow History (Filtered):
    {json.dumps(trimmed_history, indent=2)}
    """

    response = client.models.generate_content(
        model="gemini-2.5-flash",
        contents=prompt,
    )
    return response.text


def summarize_job(client, job_id, child_analyses):
    print(f"\nAnalyzing job {job_id} overall based on child analyses...")

    combined_text = "\n\n".join(
        [
            f"### Child Analysis {i + 1}\n{analysis}"
            for i, analysis in enumerate(child_analyses)
        ]
    )

    prompt = f"""
    You are an expert Principal Software Engineer and Temporal workflow architect. 
    You are tasked with analyzing the execution of a parent Job (Job ID: {job_id}) based on the individual analyses of its child workflows.
    
    Here are the individual analyses of the child workflows:
    {combined_text}
    
    Please provide a comprehensive final report with the following sections:
    
    ## 1. Overall Execution Summary
    Provide a high-level overview of what this job accomplished based on the child tasks.
    
    ## 2. Problems & Anomalies Noted
    Highlight any failures, excessive retries, or bottlenecks observed across the child workflows.
    
    ## 3. Suggested Improvements
    Suggest concrete improvements for the Temporal workflow design, activity implementations, error handling, or performance optimization.
    """

    response = client.models.generate_content(
        model="gemini-2.5-pro",
        contents=prompt,
    )
    return response.text


def main():
    parser = argparse.ArgumentParser(
        description="Analyze Temporal workflows using Gemini"
    )
    parser.add_argument(
        "--file",
        default="workflow_analysis_data.json",
        help="Path to JSON file containing workflow data",
    )
    parser.add_argument(
        "--job",
        help="Optional: Specific parent Job ID to analyze. If not provided, you will be prompted.",
    )
    parser.add_argument(
        "--children",
        type=int,
        default=3,
        help="Number of child workflows to analyze (default: 3)",
    )
    parser.add_argument(
        "--output",
        default="analysis_report.md",
        help="Output markdown file for the final report",
    )

    args = parser.parse_args()

    if not os.environ.get("GEMINI_API_KEY"):
        print("Error: GEMINI_API_KEY environment variable is required.")
        print("Please run: export GEMINI_API_KEY='your_api_key_here'")
        sys.exit(1)

    client = genai.Client()

    print(f"Loading data from {args.file}...")
    data = load_workflows(args.file)

    # Identify parent jobs (those without a parent_id)
    parent_jobs = [
        item for item in data if "parent_id" not in item or not item["parent_id"]
    ]

    if not parent_jobs:
        print("No parent jobs found in the data.")
        sys.exit(1)

    target_job = None
    if args.job:
        target_job = next(
            (j for j in parent_jobs if j["workflow_id"] == args.job), None
        )
        if not target_job:
            print(f"Job {args.job} not found.")
            sys.exit(1)
    else:
        print("\nAvailable Jobs:")
        for i, job in enumerate(parent_jobs):
            print(f"{i + 1}. {job['workflow_id']}")

        choice = int(input("\nSelect a job to analyze (number): ")) - 1
        if 0 <= choice < len(parent_jobs):
            target_job = parent_jobs[choice]
        else:
            print("Invalid choice.")
            sys.exit(1)

    job_id = target_job["workflow_id"]
    print(f"\nSelected Job: {job_id}")

    # Find children for this job
    children = [item for item in data if item.get("parent_id") == job_id]
    print(f"Found {len(children)} child workflows for this job.")

    if len(children) == 0:
        print("No child workflows found to analyze.")
        sys.exit(0)

    num_to_analyze = min(args.children, len(children))
    print(f"Analyzing {num_to_analyze} child workflows...")

    children_to_analyze = children[:num_to_analyze]

    child_analyses = []
    for child in children_to_analyze:
        analysis = analyze_child(client, child)
        child_analyses.append(analysis)

    # Generate final summary
    final_summary = summarize_job(client, job_id, child_analyses)

    # Save results
    with open(args.output, "w") as f:
        f.write(f"# Analysis Report for Job: {job_id}\n\n")

        f.write(final_summary)
        f.write("\n\n---\n\n## Individual Child Workflow Analyses\n\n")

        for i, analysis in enumerate(child_analyses):
            f.write(f"### Child {children_to_analyze[i]['workflow_id']}\n")
            f.write(analysis)
            f.write("\n\n")

    print(f"\nAnalysis complete! Report saved to {args.output}")


if __name__ == "__main__":
    main()
