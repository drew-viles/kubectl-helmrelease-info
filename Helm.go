package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"os"
	"text/tabwriter"
)

// CMNamespaceList contains a list of CMNamespaces. It is also responsible for printing the data out to the terminal.
type CMNamespaceList struct {
	Items []*CMNamespace
}

// CMNamespace containst he name of the namespace and the CMHelmRelease(s) within
type CMNamespace struct {
	Name         string
	HelmReleases []*CMHelmRelease
}

// CMHelmRelease contains the name of the release as deployed and the chart data
type CMHelmRelease struct {
	Name  string
	Chart *CMHelmChart
}

// CMHelmChart is for the actual chart data
type CMHelmChart struct {
	Name       string
	Repository string
	Version    string
	Status     string
}

// PrintChartResults will return pretty formatting of the HelmReleases
func (l *CMNamespaceList) PrintChartResults() {
	if l == nil {
		return
	}
	if l.Items == nil {
		return
	}
	if len(l.Items) <= 0 {
		return
	}

	// initialize tabwriter
	w := new(tabwriter.Writer)

	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)

	defer w.Flush()

	headers := []string{"Namespaces", "HelmRelease", "Chart", "Installed Version", "Chart Repository", "Status"}

	fmt.Fprintf(w, "\n %s", func(ch ...string) string {
		buf := ""
		for _, c := range ch {
			buf += fmt.Sprintf("%s\t\t", c)
		}
		return buf
	}(headers...))

	// prints --- as many times as there is the number of headers
	fmt.Fprintf(w, "\n %s", addSpacing(len(headers)))

	for _, ns := range l.Items {
		if ns == nil {
			continue
		}
		if ns.HelmReleases == nil {
			continue
		}
		if len(ns.HelmReleases) <= 0 {
			continue
		}
		for _, release := range ns.HelmReleases {
			{
				fmt.Fprintf(w, "\n %s", func(ch ...string) string {
					buf := ""
					for _, c := range ch {
						buf += fmt.Sprintf("%s\t\t", c)
					}
					return buf
				}([]string{ns.Name, release.Name, release.Chart.Name, release.Chart.Version, release.Chart.Repository, func(state string) string {
					if state == "Succeeded" {
						return ResourceHealthy
					}
					return ResourceUnhealthy
				}(release.Chart.Status)}...))
			}
		}
	}
	fmt.Fprintf(w, "\n")
}

// addSpacing add ---- to each col of the Table
func addSpacing(count int) interface{} {
	buf := ""
	for i := 0; i < count; i++ {
		buf += fmt.Sprintf("%s\t\t", "----")
	}
	return buf

}

// getHelmReleasesFromNamespace will set the CMNamespace.HelmReleases value with an array CMHelmRelease type
// The CMHelmRelease will be populated with a name and the chart information
func (n *CMNamespace) getHelmReleasesFromNamespace() {
	HelmReleaseRes := schema.GroupVersionResource{Group: "helm.fluxcd.io", Version: "v1", Resource: "helmreleases"}
	list, err := client.Resource(HelmReleaseRes).Namespace(n.Name).List(metav1.ListOptions{})

	if err != nil {
		log.Printf("There was an error searching for the resource:\n\t%s\n", err.Error())
		return
	}
	if len(list.Items) <= 0 {
		return
	}

	for _, release := range list.Items {
		n.HelmReleases = append(n.HelmReleases, &CMHelmRelease{
			Name:  release.GetName(),
			Chart: getHelmChartData(release),
		})
	}
}

// getHelmChartData pulls the relevant data from the HelmRelease definition and returns it as a CMHelmChart
func getHelmChartData(hr unstructured.Unstructured) *CMHelmChart {
	repository, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "repository")
	if err != nil || !found {
		log.Printf("failed to get a chart repo for: %s", hr.GetName())
	}
	name, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "name")
	if err != nil || !found {
		log.Printf("failed to get a chart name for: %s", hr.GetName())
	}
	status, found, err := unstructured.NestedString(hr.Object, "status", "phase")
	if err != nil || !found {
		log.Printf("failed to get a deployment state for: %s - cannot continue", hr.GetName())
	}
	version, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "version")
	if err != nil || !found {
		log.Printf("failed to get a chart version for: %s - cannot continue", hr.GetName())
		return nil
	}
	return &CMHelmChart{
		Name:       name,
		Version:    version,
		Repository: repository,
		Status:     status,
	}
}

// getNamespace returns all namespaces or just the one provided via flags as an array of CMNamespace type
func getNamespace() *CMNamespaceList {
	var result CMNamespaceList
	if namespace == nil || *namespace == "" {
		ns, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("There was an error getting the namespaces: %s", err)
		}
		for _, n := range ns.Items {
			result.Items = append(result.Items, &CMNamespace{
				Name: n.Name,
			})
		}
		return &result
	}
	result.Items = []*CMNamespace{{
		Name:         *namespace,
		HelmReleases: nil,
	}}
	return &result
}

// HelmReleases is the main func that will be called.
// It collect any relevant chart data from the HelmReleases and attaches it to the relevant CMNamespace
func HelmReleases() {
	namespaces := getNamespace()
	if len(namespaces.Items) <= 0 {
		log.Println("no namespaces found or passed as a flag")
		return
	}

	for _, n := range namespaces.Items {
		namespace = &n.Name
		n.getHelmReleasesFromNamespace()
	}
	namespaces.PrintChartResults()
}
