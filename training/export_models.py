"""Re-export utility for converting trained sklearn models to JSON."""
import json, os, sys, joblib

def tree_to_dict(tree, feature_names):
    t = tree.tree_
    def recurse(n):
        if t.children_left[n] == -1:
            v = t.value[n]
            if v.ndim == 2 and v.shape[1] == 2:
                total = v[0].sum()
                return {"leaf": float(v[0, 1] / total) if total > 0 else 0.0}
            return {"leaf": float(v.flat[0])}
        return {
            "feature": feature_names[int(t.feature[n])],
            "threshold": float(t.threshold[n]),
            "left": recurse(int(t.children_left[n])),
            "right": recurse(int(t.children_right[n])),
        }
    return recurse(0)

if __name__ == "__main__":
    print("Use individual training scripts to export models.")
    print("This utility is for re-exporting from saved .joblib files.")
