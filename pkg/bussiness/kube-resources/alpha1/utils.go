package alpha1

func IsHostCluster(region, cluster string) bool {
	return len(region)+len(cluster) == 0
}
