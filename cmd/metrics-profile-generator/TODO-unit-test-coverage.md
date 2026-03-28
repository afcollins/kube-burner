# Unit Test Coverage Plan

## Tier 1 (Highest Impact - Pure Logic)
1. **pkg/config** (~15 functions) - Parse, ParseWithUserdata, ParseSetValues, deepMerge, parseValue, setNestedValue, validateDNS1123, jobIsDuped, UnmarshalYAML defaults, IsChurnEnabled
2. **pkg/util** (~10 functions) - GetBoolValue, GetIntegerValue, GetStringValue, NormalizeLabels, EnvToMap, RenderTemplate, CleanupTemplate, RetryWithExponentialBackOff
3. **pkg/burner** (~8 functions) - yamlToUnstructured, yamlToUnstructuredMultiple, updateLabels, updateChildLabels, setLabelsInArray, toStatusPath, toStatusPaths
4. **pkg/measurements/metrics** (2 functions) - CheckThreshold, NewLatencySummary

## Tier 2 (Medium Impact)
5. **pkg/measurements** (~5 functions) - PodTransformFunc, createMinimalUnstructured, verifyMeasurementConfig, getIntFromLabels
6. **pkg/alerting** (2 functions) - parseMatrix, validateTemplates
7. **pkg/util/metrics** (2 functions) - DecodeMetricsEndpoint
8. **pkg/workloads** (3 functions) - extractDirectory, NewWorkloadHelper, SetVariables

## Tier 3 (External Dependencies)
9. **pkg/prometheus** - template rendering only
10. **pkg/measurements/util** - requires k8s mocks
11. **pkg/watchers** - all k8s/informer dependent
