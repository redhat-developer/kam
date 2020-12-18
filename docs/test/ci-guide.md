# ci-reference

## Running e2e on Prow
Prow is the Kubernetes or OpenShift way of managing workflow, including tests. KAM e2e test targets are passed through the script scripts/openshiftci-presubmit-all-tests.sh available in the kam repository. Prow uses the script through the command attribute of the odo job configuration file in openshift/release repository.

For running e2e test on 4.5 cluster, job configuration file will be

[source,sh]
----
    - as: integration-e2e
    steps:
        cluster_profile: aws
        test:
        - as: integration-e2e-steps
        commands: scripts/openshiftci-presubmit-all-tests.sh
        credentials:
        - mount_path: /var/run/kam-data/user-secret
            name: kam-github-secret
            namespace: test-credentials
        - mount_path: /var/run/kam-data/docker-conf
            name: kam-quay-docker-conf-secret
            namespace: test-credentials
        env:
        - default: /var/run/kam-data/user-secret/secret.txt
            name: KAM_GITHUB_TOKEN_FILE
        - default: /var/run/kam-data/docker-conf/kam-bot-kambot-auth.json
            name: KAM_QUAY_DOCKER_CONF_SECRET_FILE
        from: oc-bin-image
        resources:
            requests:
            cpu: "2"
            memory: 6Gi
        workflow: ipi-aws
----

To generate the kam job file, run make jobs in [https://github.com/openshift/release](openshift/release) for the kam pr.

Job dashboard is monitored at: [https://deck-ci.apps.ci.l2s4.p1.openshiftapps.com/?repo=redhat-developer%2Fkam](kam pr jobs dashboard)