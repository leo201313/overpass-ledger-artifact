# Artifact Description

This directory is intended for demonstrating the evaluation results of Overpass Ledger (OPL). Follow the steps below to quickly deploy OPL across multiple devices and begin testing.

---

## Step 1: Verify Environment and Files

### Environment Requirements
To enable rapid deployment of OPL, ensure you have multiple devices running Ubuntu version 20.04 or later. These devices must:
- Be accessible via SCP and SSH using a `.pem` key file.
- Have unique, independent IP addresses.

### Included Files
This folder initially contains the following files:
- **`cdntClient`**: The client used to launch the Coordinator process in OPL.
- **`workerClient`**: The client used to launch the Worker process in OPL.
- **`tmClient`**: A command-line tool for initializing OPL, running experiments, and measuring results.
- **`expInitAgent`**: A utility for generating node directories and various control scripts.
- **`initExp.yaml`**: Configuration file for initializing experiments. This file will be read by `expInitAgent`.
- **`README.md`**: This document.

### IMPORTANT
You must generate a `.pem` key file and place it in this folder. The default name for this file is `myKey.pem`.

---

## Step 2: Configure `initExp.yaml`

Edit the `initExp.yaml` file to provide necessary details, such as:
- IP addresses for each device.
- The desired number of Worker groups to deploy.

---

## Step 3: Run the Experiment Initialization Agent

Run the following command:
```bash
./expInitAgent
```

This will create the nodes directory and generate the required control scripts.

## Step 4: Deploy OPL
Execute the deployment script:
```bash
bash ./deployAll.sh
```
This will deploy OPL on all specified devices.

## Additional Instructions
After deploying OPL, it will not automatically start. To launch all OPL-related clients across devices, use the following command:
```bash
bash ./startAll.sh
```
You can then use the `tmClient` program to initialize OPL and conduct experimental tests. To learn how to use `tmClient`, execute:
```bash
./tmClient help
```

In addition, once initialization is complete, the folder also provides these control scripts:

* `stopAll.sh`: Stops all OPL-related clients on all devices.
* `resetAll.sh`: Resets all OPL-related clients on all devices to their initial state (removing all stored data).
