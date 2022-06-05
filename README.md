# Session2-kubernetes

As briefly discussed earlier, everything inside of kubernetes is described as a object. This is anything from a pod, to a piece of storage to some networking capabilities. Each one of these objects is then held in cluster and compared to the running state of the cluster to determine if any work is required. This object description is defined typically using yaml.

A yaml file for creating kubernetes objects is made up of many components that roughly split into:
1. Object type
1. Object metadata (name, location, etc)
1. Object Specification
1. Object Status

The first three are required at creation with Object status edited by a kubernetes controller. For example if we look at:
```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: kubernetes-education
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```
- The apiVersion and kind let us know exactly which object we want to create
- The metadata allows us to name the object and express which location we want it to be created in (and more)
- The spec, which is specific per object type allows us to apps configuration data to our object type.

### Setting up namespace

Before we can start setting up our application, its first good to create a dedicated namespace for our application to run in. This is a good idea, as in a production cluster there may be many applications running at the same time and we need a say of isolating them from one another, as well as controlling who has access to them.

A namespace is like any other object in kubernetes and needs to be define in yaml:
```
apiVersion: v1
kind: Namespace
metadata:
  name: kubernetes-education
```
A namespace however is one of the few objects you will likely never apply with a `spec:` define, allowing the cluster to default everything.

We can now install the yaml to our cluster using:
```
# Assuming you are in the root folder of this repo
kubectl create -f yamls/namespace.yaml 

# You can also create some objects using
kubectl apply -f yamls/namespace.yml
```
The second command will look to edit an object named data-pvc in the kubernetes-education namespace if it already exists, or create a new one if not.

We now have our own isolated workspace for out application. By default, the kubectl context is set to look at resources within the default namespace, but we can look at resources in other namespaces by either defining where we want to look:
```
kubectl get pods -n kubernetes-education
```
or by changing the context to look at a new namespace by default:
```
kubectl config set-context --current --namespace kubernetes-education
```
You can also directly edit the `~/.kube/config` file to edit the contexts

### Setting up volume

We need some small chunk of storage to be used by my applications to hold state. Remembering that all containers run in ephemeral storage, so if stopped lost. The exact mechanics how some storage is given to an application is massively varying and dependant on how a cluster is configured. So kubernetes has a layer of abstraction between the mechanism chosen to provide physical storage to an application and the user requesting it. This is called a persistentVolumeClaim. Using a PVC, we user can ask for a chunk of storage (and specify a provider is more than one), and allow kubernetes to worry about actually provisioning it. If we look at the same yaml as earlier:
```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: kubernetes-education
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```
All that has had to be concerned about is how much storage, what namespace it will be put in, and its accessModes. Under the covers the default storage class provider will complete its work to provide the request the storage and a handle to use it.

Applying the yamls as previous:
```
kubectl apply -f yamls/pvc.yml
```

Once the command is run, we can run `kubectl get pvc` to see the object we just created (pvc is the shortname for PersistentVolumeClaim)

Looking at the object from the cluster we can see it gets updated by a controller to look like this: (`kubectl edit pvc data-pvc`)
```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"PersistentVolumeClaim","metadata":{"annotations":{},"name":"data-pvc","namespace":"kubernetes-education"},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}}}
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: k8s.io/minikube-hostpath
    volume.kubernetes.io/storage-provisioner: k8s.io/minikube-hostpath
  creationTimestamp: "2022-06-05T12:43:38Z"
  finalizers:
  - kubernetes.io/pvc-protection
  name: data-pvc
  namespace: kubernetes-education
  resourceVersion: "30750"
  uid: 19ad717e-a2a5-484c-9eb1-dd3331d823eb
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
  volumeMode: Filesystem
  volumeName: pvc-19ad717e-a2a5-484c-9eb1-dd3331d823eb
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 1Gi
  phase: Bound
```
Here we can see that kubernetes has appended additional parameters (usually defaults or provisioned values) and created a status component mentioned earlier.

### Setting up configs
As an additional add on to the application previously, it can now accept startup names in a file located in `/config/defaults.json` if there is no currently stored names (aka a fresh start not restart).

This was done to show how we can pass configurations to a application, without the need to rebuild our container image. We can simply edit and pass a configuration file, using the mechanisms built into Kubernetes to store and control the configurations. This is more than we will be going into today, but there are types of applications that can be deployed that also watch objects like configMaps, and will recycle an application if something changes.

For the time being we are just going to create our config:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  namespace: kubernetes-education
data:
  defaults.json: |
    {
      "names": [
        "alan",
        "jess"
      ]
    }
```
This object represents a file called `defaults.json`, that contains:
```
{
    "names": [
    "alan",
    "jess"
    ]
}
```
You will see in the next session how we can use this for our application. Note the defaults.json file already exists in the container image ive provided, but we are looking to override it with this.

### Setting up deployment
So now we have our storage and configurations ready and waiting for us on our cluster, we have whats required to start up our app. Like the other objects, this is also defined in a yaml file:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: names
  namespace: kubernetes-education
  labels:
    app: names
spec:
  replicas: 1
  selector:
    matchLabels:
      app: names
  template:
    metadata:
      name: names
      labels:
        app: names
    spec:
      initContainers:
      - name: init-data
        image: busybox:1.28
        command: ['sh', '-c', 'chown 200 /data']
        volumeMounts:
        - name: datadir
          mountPath: /data
      containers:
      - name: names
        image: docker.io/library/namesapp:v1
        ports:
        - containerPort: 8776
          name: http
        volumeMounts:
        - name: datadir
          mountPath: /data
        - name: defaults
          mountPath: /config
        resources:
          requests:
            cpu: "4"
      volumes:
      - name: defaults
        configMap:
          name: config
      - name: datadir
        persistentVolumeClaim:
          claimName: data-pvc
```
But this is a little more complicated, with a couple of components defined. This yaml is actually for defining a deployment, which is an abstraction above just a pod. I could define a yaml that describes my runtime as a pod, but this would hold now application management levels of logic, as well as not holding a state on the deployment of our app. A single defined pod, is just that, and if ever lost it would not be restarted. This is because as soon as the object is not running, the object is recycled, clearing the state and spec with it.

So we can use something like a deployment to define our application as a set of pods. That is why when we look at the `spec:`, there is a `template:` with its own `spec: `defined. This is because this template is the template for our pod, which kubernetes will try to maintain even if it is crashing or is lost. Like other components, the template is just one of the defining pieces for a deployment, but resulting in the creation of a secondary child object. This is a good point to mention that if a parent object is deleted, so are its children (unless explicitly expressed).

Breaking up some of the more important components of this:
```
labels:
  app: names
```
Labels can be applied to objects to help select them later.

```
initContainers:
- name: init-data
  image: busybox:1.28
  command: ['sh', '-c', 'chown 200 /data']
  volumeMounts:
  - name: datadir
    mountPath: /data
```
Like in the compile dockerfile from last session, this is a secondary container that is brought up in our pod to perform some setup work, in this case providing file ownership to the mounted storage.

```
- name: names
  image: docker.io/library/namesapp:v1
  ports:
  - containerPort: 8766
    name: http
    volumeMounts:
    - name: datadir
      mountPath: /data
    - name: defaults
      mountPath: /config
```
Describes out runtime container. There is many options that can be passed here, including environment variables and runtime parameters. Here though we are defining a port to open on the pod, and two mountpoints for volumes.
```
volumes:
- name: defaults
  configMap:
    name: config
- name: datadir
  persistentVolumeClaim:
    claimName: data-pvc
```

`defaults` is the config map we created a second ago containing our json file.
`datadir` is the volume created from the PVC and our persistent storage. 

These volumes then get mounted onto the volume mounts defined in the `containers:`

Now applying this:
```
kubectl apply -f yamls/deployments.yaml
```
Will go and start up our containers. Use `kubectl get pods` to look at the pod, after a few seconds it should go to a state of `running`.

### Exposing our application

Now we have our application running, how do we access it? By default each namespace is in its own subnet and doesnt expose anything. So we now have to expose this applcation with a final object called a service:
```
apiVersion: v1
kind: Service
metadata:
  name: names-external
  namespace: kubernetes-education
spec:
  type: NodePort
  ports:
  - name: http
    port: 8766
    protocol: TCP
    targetPort: 8766
    # We dont usually set this
    nodePort: 31463
  selector:
    app: names
```
In this example we are using a NodePort type, which is the simplest to get started with, as it physically opens up a port on the control-plane node to be directed at our application. There are more mechanisms like loadBalancers and ingresses which can also be used.

As I have put, we dont usually set the nodePort itself, as we let kubernetes retrieve a free one, rather than choosing. I have only done this as when we started the minikube cluster, we specifically opened this port for the WSL2 windows users. If on a mac feel free to delete this line and let kubernetes provision a port for you.

Also in this service we use a selector to find runtimes labeled with `app:names`.

If we apply this yaml:
```
kubectl apply -f yamls/svc.yaml
```
We should then be able to access our application through `localhost:31463` (or what ever port was provisioned for you `kubectl get svc` to see)