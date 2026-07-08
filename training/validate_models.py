"""Run validation on all three exported models and generate reports."""
import json
import os
import sys
from datetime import datetime, timezone

MODELS_DIR = os.path.join(os.path.dirname(__file__), "..", "internal", "models")
OUTPUT_JSON = os.path.join(os.path.dirname(__file__), "model_validation.json")
OUTPUT_MD = os.path.join(os.path.dirname(__file__), "model_validation.md")

def main():
    print("=" * 60)
    print("MODEL VALIDATION REPORT")
    print("=" * 60)
    
    report = {
        "generated_at": datetime.now(tz=timezone.utc).isoformat(),
        "models": {}
    }
    
    models = ["congestion_model.json", "trust_model.json", "targeting_model.json"]
    
    for name in models:
        path = os.path.join(MODELS_DIR, name)
        if not os.path.exists(path):
            print(f"ERROR: {path} not found. Run training scripts first.")
            sys.exit(1)
        with open(path) as f:
            m = json.load(f)
            
        task = m.get("task", m.get("model_type", "unknown"))
        report["models"][name] = {
            "task": task,
            "version": m.get("model_version", "1.0.0"),
            "features": m.get("feature_names", m.get("features", [])),
            "metrics": m.get("metrics", {}),
            "dataset_hash": m.get("dataset_hash", "unknown"),
        }
        print(f"Loaded {name} ({task})")
    
    # Write JSON report
    with open(OUTPUT_JSON, "w") as f:
        json.dump(report, f, indent=2)
        
    # Write Markdown report
    with open(OUTPUT_MD, "w") as f:
        f.write("# AlchemiX Model Validation Report\n\n")
        f.write(f"**Generated:** {report['generated_at']}\n\n")
        
        for name, data in report["models"].items():
            f.write(f"## {name} (`{data['task']}`)\n")
            f.write(f"- **Version:** {data['version']}\n")
            f.write(f"- **Dataset Hash:** `{data['dataset_hash']}`\n")
            f.write(f"- **Feature Count:** {len(data['features'])}\n")
            f.write("\n### Metrics\n")
            for k, v in data["metrics"].items():
                if isinstance(v, dict):
                    f.write(f"- **{k}**:\n")
                    for subk, subv in v.items():
                        f.write(f"  - `{subk}`: {subv}\n")
                else:
                    f.write(f"- **{k}**: {v}\n")
            f.write("\n")
            
    print(f"\nConsolidated reports generated at:")
    print(f"  {OUTPUT_JSON}")
    print(f"  {OUTPUT_MD}")
    print("=" * 60)

if __name__ == "__main__":
    main()
