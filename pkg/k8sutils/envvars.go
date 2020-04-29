package k8sutils

import (
	corev1 "k8s.io/api/core/v1"
)

// MergeEnvVars merges env variables by name
func MergeEnvVars(envs []corev1.EnvVar, additionalEnvs []corev1.EnvVar) []corev1.EnvVar {
	if len(additionalEnvs) == 0 {
		return envs
	}

	indexedByName := make(map[string]int)
	variables := make([]corev1.EnvVar, len(envs))

	for i, env := range envs {
		indexedByName[env.Name] = i
		variables[i] = env
	}

	for _, env := range additionalEnvs {
		if idx, ok := indexedByName[env.Name]; ok {
			variables[idx] = env
		} else {
			variables = append(variables, env)
		}
	}

	return variables
}
