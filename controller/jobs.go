package controller

import (
	"context"
	"fmt"
	v12 "k8s.io/api/batch/v1"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	_ "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"lightspeed-bot-controller/persistence"
	"log"
	"os"
	"time"
)

const (
	namespace = "lightspeed-bot"
)

var client *kubernetes.Clientset
var ttlAfterFinish int32 = 10

var jobsInformer cache.SharedIndexInformer

func CreateKubernetesClient(ctx context.Context) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBERNETES_CONFIG"))
	if err != nil {
		panic(err)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour,
		informers.WithNamespace("lightspeed-bot"))
	jobsInformer = factory.Batch().V1().Jobs().Informer()
	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	log.Println("informer cache synced")
}

func RunListener(ctx context.Context) {
	jobsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		UpdateFunc: func(oldObj, newObj interface{}) {
			job := newObj.(*v12.Job)
			if _, ok := job.Labels["role"]; !ok {
				return
			}

			for _, condition := range job.Status.Conditions {
				if condition.Type == v12.JobFailed {
					log.Println(fmt.Sprintf("job %s for %s has failed",
						job.Name, job.Labels["repository"]))
					return
				}
				if condition.Type == v12.JobComplete {
					log.Println(fmt.Sprintf("job %s for %s has completed",
						job.Name, job.Labels["repository"]))
					return
				}
			}
		},
		DeleteFunc: func(job interface{}) {
			deletedJob := job.(*v12.Job)
			log.Println(fmt.Sprintf("job %s deleted", deletedJob.Name))
		},
	})
}

func IsScanInProgress(ctx context.Context, repositoryID int64) (bool, error) {
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("repository=%d", repositoryID),
	}
	jobs, err := client.BatchV1().Jobs(namespace).List(ctx, opts)
	if err != nil {
		return false, err
	}

	return len(jobs.Items) > 0, nil
}

func CreateScanJob(ctx context.Context, id int64, repo string, owner string, token string) error {
	var backoffLimit int32 = 1
	job := &v12.Job{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("scan-%d", id),
			Labels: map[string]string{
				"repository": fmt.Sprintf("%d", id),
				"role":       "ansible-lightspeed-scan-job",
			},
			Namespace: namespace,
		},
		Spec: v12.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlAfterFinish,
			Template: v13.PodTemplateSpec{
				Spec: v13.PodSpec{
					Containers: []v13.Container{
						{
							Name:  "content-scanner",
							Image: "quay.io/pb82/content-scanner:latest",
							Env: []v13.EnvVar{
								{
									Name:  "REPOSITORY",
									Value: repo,
								},
								{
									Name:  "OWNER",
									Value: owner,
								},
								{
									Name:  "GITHUB_TOKEN",
									Value: token,
								},
							},
							ImagePullPolicy: v13.PullAlways,
						},
					},
					RestartPolicy: v13.RestartPolicyNever,
				},
			},
		},
	}

	_, err := client.BatchV1().Jobs(namespace).Create(ctx, job, v1.CreateOptions{})
	if err == nil {
		persistence.UpdateLastScanned(id)
	}

	return err
}
