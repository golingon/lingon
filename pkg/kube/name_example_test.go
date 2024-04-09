// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"fmt"

	"github.com/golingon/lingon/pkg/kubeutil"
)

func ExampleNameVarFunc_addKind() {
	m := kubeutil.Metadata{
		Kind: "Deployment",
		Meta: kubeutil.Meta{Name: "super-duper-app"},
	}
	fmt.Println(NameVarFunc(m))
	// Output: SuperDuperAppDeploy
}

func ExampleNameVarFunc_kindSuffix() {
	m := kubeutil.Metadata{
		Kind: "Deployment",
		Meta: kubeutil.Meta{Name: "super-duper-deployment"},
	}
	fmt.Println(NameVarFunc(m))
	// Output: SuperDuperDeploy
}

func ExampleNameVarFunc_kindWithDash() {
	m := kubeutil.Metadata{
		Kind: "ClusterRole",
		Meta: kubeutil.Meta{Name: "argo-cluster-role"},
	}
	fmt.Println(NameVarFunc(m))
	// Output: ArgoCR
}

func ExampleNameFieldFunc_addKind() {
	m := kubeutil.Metadata{
		Kind: "Deployment",
		Meta: kubeutil.Meta{Name: "super-duper-app"},
	}
	fmt.Println(NameFieldFunc(m))
	// Output: SuperDuperAppDeploy
}

func ExampleNameFieldFunc_kindSuffix() {
	m := kubeutil.Metadata{
		Kind: "Deployment",
		Meta: kubeutil.Meta{Name: "super-duper-deployment"},
	}
	fmt.Println(NameFieldFunc(m))
	// Output: SuperDuperDeploy
}

func ExampleNameFieldFunc_kindWithDash() {
	m := kubeutil.Metadata{
		Kind: "ClusterRole",
		Meta: kubeutil.Meta{Name: "argo-cluster-role"},
	}
	fmt.Println(NameFieldFunc(m))
	// Output: ArgoCR
}

func ExampleNameFileFunc() {
	m := kubeutil.Metadata{
		Kind: "Deployment",
		Meta: kubeutil.Meta{Name: "super-duper-app"},
	}
	fmt.Println(NameFileFunc(m))
	// Output: super-duper-app_deploy.go
}

func ExampleDirectoryName_defaultNamespace() {
	fmt.Println(DirectoryName("myapps", "Deployment"))
	// Output: myapps
}

func ExampleDirectoryName_clusterrole() {
	fmt.Println(DirectoryName("", "ClusterRole"))
	// Output: _cluster/rbac
}

func ExampleDirectoryName_crd() {
	fmt.Println(DirectoryName("", "CustomResourceDefinition"))
	// Output: _cluster/crd
}
