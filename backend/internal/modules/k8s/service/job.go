package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type JobListVO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Completions     int32             `json:"completions"`
	Parallelism     int32             `json:"parallelism"`
	Active          int32             `json:"active"`
	Succeeded       int32             `json:"succeeded"`
	Failed          int32             `json:"failed"`
	Labels          map[string]string `json:"labels"`
	Containers      []ContainerInfo   `json:"containers"`
	ResourceSummary ResourceSummary   `json:"resourceSummary"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"createdAt"`
	CompletionTime  *time.Time        `json:"completionTime,omitempty"`
	Duration        string            `json:"duration,omitempty"`
}

type JobListResponse struct {
	Total int64       `json:"total"`
	Items []JobListVO `json:"items"`
}

func (s *K8sService) ListJobs(clusterName string, namespace string, page, pageSize int, keyword string) (*JobListResponse, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	list, err := client.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(cluster.Name)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item batchv1.Job) []string {
		images := make([]string, 0, len(item.Spec.Template.Spec.Containers))
		for _, container := range item.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
		return []string{
			item.Name,
			item.Namespace,
			strings.Join(images, ","),
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	result := make([]JobListVO, 0, len(paged))
	for _, item := range paged {
		containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
		resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

		vo := JobListVO{
			Name:            item.Name,
			Namespace:       item.Namespace,
			Completions:     derefInt32(item.Spec.Completions),
			Parallelism:     derefInt32(item.Spec.Parallelism),
			Active:          item.Status.Active,
			Succeeded:       item.Status.Succeeded,
			Failed:          item.Status.Failed,
			Labels:          item.Labels,
			Containers:      containers,
			ResourceSummary: resSummary,
			Status:          computeJobStatus(&item),
			CreatedAt:       item.CreationTimestamp.Time,
			CompletionTime:  nil,
			Duration:        "",
		}
		if item.Status.CompletionTime != nil {
			vo.CompletionTime = &item.Status.CompletionTime.Time
			vo.Duration = formatDuration(item.CreationTimestamp.Time, item.Status.CompletionTime.Time)
		}
		result = append(result, vo)
	}

	return &JobListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetJobDetail(clusterName string, namespace, name string) (*JobListVO, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.BatchV1().Jobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
	resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

	vo := &JobListVO{
		Name:            item.Name,
		Namespace:       item.Namespace,
		Completions:     derefInt32(item.Spec.Completions),
		Parallelism:     derefInt32(item.Spec.Parallelism),
		Active:          item.Status.Active,
		Succeeded:       item.Status.Succeeded,
		Failed:          item.Status.Failed,
		Labels:          item.Labels,
		Containers:      containers,
		ResourceSummary: resSummary,
		Status:          computeJobStatus(item),
		CreatedAt:       item.CreationTimestamp.Time,
		CompletionTime:  nil,
		Duration:        "",
	}
	if item.Status.CompletionTime != nil {
		vo.CompletionTime = &item.Status.CompletionTime.Time
		vo.Duration = formatDuration(item.CreationTimestamp.Time, item.Status.CompletionTime.Time)
	}
	return vo, nil
}

func (s *K8sService) GetJobYAML(clusterName string, namespace, name string) (string, error) {
	obj, err := s.GetJobObject(clusterName, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) GetJobObject(clusterName string, namespace, name string) (*batchv1.Job, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.BatchV1().Jobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) CreateJob(clusterName string, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.BatchV1().Jobs(namespace).Create(context.Background(), job, metav1.CreateOptions{})
}

func (s *K8sService) DeleteJob(clusterName string, namespace, name string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.BatchV1().Jobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func computeJobStatus(job *batchv1.Job) string {
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed && cond.Status == "True" {
			return "Failed"
		}
		if cond.Type == batchv1.JobComplete && cond.Status == "True" {
			return "Completed"
		}
	}
	if job.Status.Active > 0 {
		return "Running"
	}
	return "Pending"
}

func formatDuration(start, end time.Time) string {
	d := end.Sub(start)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
