package k8sutils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	redisv1 "redis-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RedisDetails will hold the information for Redis Pod
type RedisDetails struct {
	PodName   string
	Namespace string
}

// GetRedisServerIP will return the IP of redis service
func GetRedisServerIP(redisInfo RedisDetails) string {
	reqLogger := log.WithValues("Request.Namespace", redisInfo.Namespace, "Request.PodName", redisInfo.PodName)
	redisIP, _ := GenerateK8sClient().CoreV1().Pods(redisInfo.Namespace).
		Get(context.TODO(), redisInfo.PodName, metav1.GetOptions{})

	reqLogger.Info("Successfully got the ip for redis", "ip", redisIP.Status.PodIP, "hostIp", redisIP.Status.HostIP)
	return redisIP.Status.PodIP
}

// ExecuteRedisClusterCommand will execute redis cluster creation command
func ExecuteRedisClusterCommand(cr *redisv1.Redis) error {
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	replicas := cr.Spec.Size
	cmd := []string{
		"redis-cli",
		"--cluster",
		"create",
	}
	for podCount := 0; podCount <= int(*replicas)-1; podCount++ {
		pod := RedisDetails{
			PodName:   cr.ObjectMeta.Name + "-master-" + strconv.Itoa(podCount),
			Namespace: cr.Namespace,
		}
		cmd = append(cmd, GetRedisServerIP(pod)+":6379")
	}
	cmd = append(cmd, "--cluster-yes")
	if cr.Spec.GlobalConfig.Password != nil {
		cmd = append(cmd, "-a")
		cmd = append(cmd, *cr.Spec.GlobalConfig.Password)
	}
	reqLogger.Info("Redis cluster creation command is", "Command", cmd)
	return ExecuteCommand(cr, cmd)
}

// CreateRedisReplicationCommand will create redis replication creation command
func CreateRedisReplicationCommand(cr *redisv1.Redis, nodeNumber string) []string {
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	cmd := []string{
		"redis-cli",
		"--cluster",
		"add-node",
	}
	masterPod := RedisDetails{
		PodName:   cr.ObjectMeta.Name + "-master-" + nodeNumber,
		Namespace: cr.Namespace,
	}
	slavePod := RedisDetails{
		PodName:   cr.ObjectMeta.Name + "-slave-" + nodeNumber,
		Namespace: cr.Namespace,
	}
	cmd = append(cmd, GetRedisServerIP(slavePod)+":6379")
	cmd = append(cmd, GetRedisServerIP(masterPod)+":6379")
	cmd = append(cmd, "--cluster-slave")

	if cr.Spec.GlobalConfig.Password != nil {
		cmd = append(cmd, "-a")
		cmd = append(cmd, *cr.Spec.GlobalConfig.Password)
	}
	reqLogger.Info("Redis replication creation command is", "Command", cmd)
	return cmd
}

// ExecuteRedisReplicationCommand will execute the replication command
func ExecuteRedisReplicationCommand(cr *redisv1.Redis) error {
	replicas := cr.Spec.Size
	for podCount := 0; podCount <= int(*replicas)-1; podCount++ {
		cmd := CreateRedisReplicationCommand(cr, strconv.Itoa(podCount))
		err := ExecuteCommand(cr, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckRedisCluster will check the redis cluster have sufficient nodes or not
func CheckRedisCluster(cr *redisv1.Redis) int {
	var client *redis.Client
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)

	redisInfo := RedisDetails{
		PodName:   cr.ObjectMeta.Name + "-master-0",
		Namespace: cr.Namespace,
	}

	if cr.Spec.GlobalConfig.Password != nil {
		//client = redis.NewClient(&redis.Options{
		//	Addr:     "127.0.0.1:6379",
		//	Password: *cr.Spec.GlobalConfig.Password,
		//	DB:       0,
		//})
		client = redis.NewClient(&redis.Options{
			Addr:     GetRedisServerIP(redisInfo) + ":6379",
			Password: *cr.Spec.GlobalConfig.Password,
			DB:       0,
		})
	} else {
		//client = redis.NewClient(&redis.Options{
		//	Addr:     "127.0.0.1:6379",
		//	Password: "Opstree@1234",
		//	DB:       0,
		//})
		client = redis.NewClient(&redis.Options{
			Addr:     GetRedisServerIP(redisInfo) + ":6379",
			Password: "",
			DB:       0,
		})
	}
	cmd := redis.NewStringCmd("cluster", "nodes")
	fmt.Println("######## CheckRedisCluster cmd: ", cmd)
	err := client.Process(cmd)
	if err != nil {
		reqLogger.Error(err, "Redis command failed with this error")
	}

	output, err := cmd.Result()
	if err != nil {
		reqLogger.Error(err, "Redis command failed with this error")
	}
	reqLogger.Info("Redis cluster nodes are listed", "Output", output)
	scanner := bufio.NewScanner(strings.NewReader(output))

	count := 0
	for scanner.Scan() {
		count++
	}
	reqLogger.Info("Total number of redis nodes are", "Nodes", strconv.Itoa(count))
	return count
}

// ExecuteCommand will execute the commands in pod
func ExecuteCommand(cr *redisv1.Redis, cmd []string) error {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	//config, _ := rest.InClusterConfig()
	//config, _ := clientcmd.BuildConfigFromFlags("", "/Users/shifu/.kube/config")
	config := ctrl.GetConfigOrDie()

	pod, err := GenerateK8sClient().CoreV1().Pods(cr.Namespace).Get(context.TODO(), cr.ObjectMeta.Name+"-master-0", metav1.GetOptions{})

	if err != nil {
		reqLogger.Error(err, "Could not get pod info")
	}

	targetContainer := -1
	for i, tr := range pod.Spec.Containers {
		reqLogger.Info("Pod Counted successfully", "Count", i, "Container Name", tr.Name)
		if tr.Name == cr.ObjectMeta.Name+"-master" {
			targetContainer = i
			break
		}
	}

	if targetContainer < 0 {
		reqLogger.Error(err, "Could not find pod to execute")
	}

	req := GenerateK8sClient().CoreV1().RESTClient().Post().
		Resource("pods").
		Name(cr.ObjectMeta.Name + "-master-0").
		Namespace(cr.Namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[targetContainer].Name,
		Command:   cmd,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		reqLogger.Error(err, "Failed to init executor")
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    false,
	})

	if err != nil {
		reqLogger.Error(err, "Could not execute command")
	}
	reqLogger.Info("Successfully executed the command", "Command", cmd, "Output", execOut.String())
	return err
}
