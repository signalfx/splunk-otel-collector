"""Creates metrics yaml section for openshift monitor"""
import yaml
import sys

RESOURCES = {
    "cpu": "number",
    "memory": "amount",
    "pods": "number",
    "services": "number",
    "persistentvolumeclaims": "number",
    "services.nodeports": "number",
    "services.loadbalancers": "number",
}

METRICS = {}


def build_metrics(metricPrefix, end):
    for resource, typ in RESOURCES.items():
        for category in ("hard", "used"):
            if category == "hard":
                description = f"Hard limit for {typ} of {resource} {end}"
            else:
                description = f"Consumed {typ} of {resource} {end}"

            yield (f"{metricPrefix}.{resource}.{category}", {
                "description": description,
                "included": False,
                "type": "gauge",
            })

METRICS.update(build_metrics("openshift.clusterquota", "across all namespaces"))
METRICS.update(build_metrics("openshift.appliedclusterquota", "by namespace"))

yaml.safe_dump({"metrics": METRICS}, sys.stdout, encoding='utf-8', default_flow_style=False)
