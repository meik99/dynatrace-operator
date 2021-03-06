curl -s "https://raw.githubusercontent.com/ \
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash

template_image="dynatrace-operator:snapshot"
current_image="dynatrace-operator:${TRAVIS_TAG}"
mkdir artefacts

kustomize build github.com/Dynatrace/dynatrace-operator/deploy/kubernetes -o kubernetes.yaml
kustomize build github.com/Dynatrace/dynatrace-operator/deploy/openshift -o openshift.yaml

sed "s/${template_image}/${current_image}/g" kubernetes.yaml >artefacts/kubernetes.yaml
sed "s/docker.io\/dynatrace\/${template_image}/registry.connect.redhat.com\/dynatrace\/${current_image}/g" openshift.yaml >artefacts/openshift.yaml
