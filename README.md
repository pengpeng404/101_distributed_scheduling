# Distributed Scheduling


Following Kubernetes best practices,
we aim for the applications deployed in our cluster to be as stateless as possible.
Thus, stateful applications and single workload applications are not suitable for deployment on cloud service providers' spot instances.
This plan involves only Deployments;
for StatefulSet applications,
they cannot be scheduled simply through scheduling policies but should be designed with the involvement of dedicated technical experts.




## Analyze



We hope to achieve our goals with minimal changes,
so modifying the scheduler rules is not an option.
Therefore, using an admission plugin is a good choice.
Configuring an appropriate mutating webhook to apply an affinity to the deployment can schedule the application to spot instances.

- nodeAffinity
  - Schedule the application to nodes with the 'node.kubernetes.io/capacity: spot' label
- podAntiAffinity
  - Distribute pods across different availability zones, different instance types, and different instances. This can maximize the dispersion of pods for the same service, avoiding service downtime due to interruptions in spot instances










