package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

type CronJobListVO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Schedule        string            `json:"schedule"`
	Suspend         *bool             `json:"suspend"`
	ActiveJobs      int32             `json:"activeJobs"`
	LastSchedule    *time.Time        `json:"lastSchedule,omitempty"`
	Labels          map[string]string `json:"labels"`
	Containers      []ContainerInfo   `json:"containers"`
	ResourceSummary ResourceSummary   `json:"resourceSummary"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"createdAt"`
}

type CronJobListResponse struct {
	Total int64          `json:"total"`
	Items []CronJobListVO `json:"items"`
}

func (s *K8sService) ListCronJobs(clusterName string, namespace string, page, pageSize int, keyword string) (*CronJobListResponse, error) {
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

	list, err := client.BatchV1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(cluster.Name)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item batchv1.CronJob) []string {
		images := make([]string, 0, len(item.Spec.JobTemplate.Spec.Template.Spec.Containers))
		for _, container := range item.Spec.JobTemplate.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
		return []string{
			item.Name,
			item.Namespace,
			item.Spec.Schedule,
			strings.Join(images, ","),
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	result := make([]CronJobListVO, 0, len(paged))
	for _, item := range paged {
		containers := buildContainerInfos(item.Spec.JobTemplate.Spec.Template.Spec.Containers)
		resSummary := aggregateResources(item.Spec.JobTemplate.Spec.Template.Spec.Containers)

		vo := CronJobListVO{
			Name:            item.Name,
			Namespace:       item.Namespace,
			Schedule:        item.Spec.Schedule,
			Suspend:         item.Spec.Suspend,
			ActiveJobs:      int32(len(item.Status.Active)),
			LastSchedule:    nil,
			Labels:          item.Labels,
			Containers:      containers,
			ResourceSummary: resSummary,
			Status:          computeCronJobStatus(&item),
			CreatedAt:       item.CreationTimestamp.Time,
		}
		if item.Status.LastScheduleTime != nil {
			vo.LastSchedule = &item.Status.LastScheduleTime.Time
		}
		result = append(result, vo)
	}

	return &CronJobListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetCronJobDetail(clusterName string, namespace, name string) (*CronJobListVO, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.BatchV1().CronJobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := buildContainerInfos(item.Spec.JobTemplate.Spec.Template.Spec.Containers)
	resSummary := aggregateResources(item.Spec.JobTemplate.Spec.Template.Spec.Containers)

	vo := &CronJobListVO{
		Name:            item.Name,
		Namespace:       item.Namespace,
		Schedule:        item.Spec.Schedule,
		Suspend:         item.Spec.Suspend,
		ActiveJobs:      int32(len(item.Status.Active)),
		LastSchedule:    nil,
		Labels:          item.Labels,
		Containers:      containers,
		ResourceSummary: resSummary,
		Status:          computeCronJobStatus(item),
		CreatedAt:       item.CreationTimestamp.Time,
	}
	if item.Status.LastScheduleTime != nil {
		vo.LastSchedule = &item.Status.LastScheduleTime.Time
	}
	return vo, nil
}

func (s *K8sService) GetCronJobYAML(clusterName string, namespace, name string) (string, error) {
	obj, err := s.GetCronJobObject(clusterName, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) GetCronJobObject(clusterName string, namespace, name string) (*batchv1.CronJob, error) {
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
	return client.BatchV1().CronJobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) CreateCronJob(clusterName string, namespace string, cj *batchv1.CronJob) (*batchv1.CronJob, error) {
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

	return client.BatchV1().CronJobs(namespace).Create(context.Background(), cj, metav1.CreateOptions{})
}

func (s *K8sService) UpdateCronJob(clusterName string, namespace string, cj *batchv1.CronJob) (*batchv1.CronJob, error) {
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

	return client.BatchV1().CronJobs(namespace).Update(context.Background(), cj, metav1.UpdateOptions{})
}

func (s *K8sService) UpdateCronJobByYAML(clusterName string, namespace, name, rawYAML string) (*batchv1.CronJob, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetCronJobObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current cronjob: %w", err)
	}

	var desired batchv1.CronJob
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = batchv1.CronJobStatus{}
	desired.ManagedFields = nil

	return s.UpdateCronJob(clusterName, namespace, &desired)
}

func (s *K8sService) SuspendCronJob(clusterName string, namespace, name string, suspend bool) (*batchv1.CronJob, error) {
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

	patch := fmt.Sprintf(`{"spec":{"suspend":%v}}`, suspend)
	return client.BatchV1().CronJobs(namespace).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
}

func (s *K8sService) DeleteCronJob(clusterName string, namespace, name string) error {
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

	return client.BatchV1().CronJobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func computeCronJobStatus(cj *batchv1.CronJob) string {
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return "Suspended"
	}
	if len(cj.Status.Active) > 0 {
		return "Active"
	}
	if cj.Status.LastScheduleTime != nil {
		return "Scheduled"
	}
	return "Pending"
}
