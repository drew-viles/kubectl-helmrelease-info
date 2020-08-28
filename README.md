# kubectl-helmrelease-info

*Simple kubectl plugin*

If you're using [krew](https://github.com/kubernetes-sigs/krew), download and place in your /path/to/.krew/bin

If you're not using [krew](https://github.com/kubernetes-sigs/krew) then place the binary in /usr/local/bin as described [here](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#using-a-plugin).

## Version
V0.1.0
 
## Features
  * Returns current serverversion
  * Returns Helm Releases in formatted table with: 
    * Chart name
    * Chart version
    * Source repo
    * Deployment Status 
    
## Usage

 ```
kubectl helmrelease info

  -kubeconfig string
        (optional) absolute path to the kubeconfig file (default "/home/drew/.kube/config")
  -n string
        specify the namespace to get the helm release data from
 ```

## Upcoming plans
I'll be looking at letting the user know if a chart is out of date where possible.
since all Helm chart sources are different, this may prove to be a problem.