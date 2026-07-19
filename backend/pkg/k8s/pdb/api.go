package pdb

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetPDBList(client *kubernetes.Clientset, namespace string) ([]policyv1.PodDisruptionBudget, error) {
	pdbList, err := client.PolicyV1().PodDisruptionBudgets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pdbList.Items, nil
}

func GetPDBYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	pdb, err := client.PolicyV1().PodDisruptionBudgets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	pdb.TypeMeta = metav1.TypeMeta{APIVersion: "policy/v1", Kind: "PodDisruptionBudget"}
	out, err := yamlutil.MarshalWithoutManagedFields(pdb)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PDB to YAML: %w", err)
	}
	return string(out), nil
}

func CreatePDB(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var pdb policyv1.PodDisruptionBudget
	if err := yaml.Unmarshal([]byte(yamlContent), &pdb); err != nil {
		return fmt.Errorf("failed to unmarshal PDB YAML: %w", err)
	}
	if pdb.Namespace == "" {
		pdb.Namespace = namespace
	}
	_, err := client.PolicyV1().PodDisruptionBudgets(namespace).Create(context.TODO(), &pdb, metav1.CreateOptions{})
	return err
}

func UpdatePDB(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var pdb policyv1.PodDisruptionBudget
	if err := yaml.Unmarshal([]byte(yamlContent), &pdb); err != nil {
		return fmt.Errorf("failed to unmarshal PDB YAML: %w", err)
	}
	if pdb.Namespace == "" {
		pdb.Namespace = namespace
	}
	_, err := client.PolicyV1().PodDisruptionBudgets(namespace).Update(context.TODO(), &pdb, metav1.UpdateOptions{})
	return err
}

func DeletePDB(client *kubernetes.Clientset, namespace, name string) error {
	return client.PolicyV1().PodDisruptionBudgets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func GetPDBDetail(client *kubernetes.Clientset, namespace, name string) (*policyv1.PodDisruptionBudget, error) {
	return client.PolicyV1().PodDisruptionBudgets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
