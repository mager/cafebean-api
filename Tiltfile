docker_build('caffy-beans-api', '.', dockerfile='Dockerfile')
k8s_yaml('deployment.yaml')
k8s_resource('caffy-beans-api', port_forwards=8080)