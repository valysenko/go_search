import requests

response = requests.get(url='http://localhost:8083/connectors?expand=info&expand=status')
data = response.json()

for connector_name, connector in data.items():
    tasks = connector['status']['tasks']
    if len(tasks) < 1:
        continue
    for task in tasks:
        if task['state'] != "RUNNING":
            print(connector_name, task)
            exit(1)
