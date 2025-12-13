Grafana Canvas
==============

1.	namespace を作成する

	```bash
	kubectl create ns boids-simu

	1. secret を作成する

	```

	bash kubectl create secret generic grafana-url -n boids-simu --from-literal=url=$GRAFANA_URL kubectl create secret generic grafana-token -n boids-simu --from-literal=url=$GRAFANA_TOKEN

	```

	1. アプリをデプロイ

	```

	bash kubectl apply -f manifest/sync/deployment.yaml\`\`\`

	参考
	----

	-	[【フリー素材】宇宙の背景イラスト](https://aipict.com/utility_graphics/universe/)
