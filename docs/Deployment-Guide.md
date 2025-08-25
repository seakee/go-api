# Deployment Guide | 部署指南

[English](#english) | [中文](#中文)

---

## English

### Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Linux/macOS/Windows server
- Minimum 1GB RAM, 1 CPU core
- Network access for database and Redis

### Environment Setup

#### 1. Production Configuration

Create production configuration file:

```bash
# Create production config
cp bin/configs/local.json.default bin/configs/prod.json
```

Edit `bin/configs/prod.json`:

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "release",
    "http_port": ":8080",
    "read_timeout": 60,
    "write_timeout": 60,
    "jwt_secret": "your-production-secret-key-change-this",
    "default_lang": "en-US"
  },
  "log": {
    "driver": "file",
    "level": "info",
    "path": "/bin/logs/"
  },
  "databases": [
    {
      "enable": true,
      "db_type": "mysql",
      "db_name": "go_api_prod",
      "db_host": "mysql:3306",
      "db_username": "go_api_user",
      "db_password": "secure_password_here",
      "db_max_idle_conn": 10,
      "db_max_open_conn": 100,
      "db_max_lifetime": 3
    }
  ],
  "redis": [
    {
      "enable": true,
      "name": "go-api",
      "host": "redis:6379",
      "auth": "redis_password_here",
      "max_idle": 30,
      "max_active": 100,
      "idle_timeout": 30,
      "prefix": "go-api-prod",
      "db": 0
    }
  ],
  "kafka": {
    "brokers": ["kafka1:9092", "kafka2:9092", "kafka3:9092"],
    "max_retry": 3,
    "client_id": "go-api-prod",
    "producer_enable": true,
    "consumer_enable": true,
    "consumer_group": "go-api-consumer-group",
    "consumer_topics": ["events", "notifications"],
    "consumer_auto_submit": true
  },
  "monitor": {
    "panic_robot": {
      "enable": true,
      "feishu": {
        "enable": true,
        "push_url": "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
      }
    }
  }
}
```

#### 2. Environment Variables

Set required environment variables:

```bash
export RUN_ENV=prod
export APP_NAME=go-api
export JWT_SECRET=your-production-secret-key
export DB_PASSWORD=secure_database_password
export REDIS_PASSWORD=secure_redis_password
```

### Docker Deployment

#### 1. Docker Compose (Recommended)

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  go-api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: go-api-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./bin/configs:/bin/configs:ro
      - ./bin/logs:/bin/logs
      - ./bin/data:/bin/data:ro
    environment:
      - RUN_ENV=prod
      - APP_NAME=go-api
      - TZ=Asia/Shanghai
    depends_on:
      - mysql
      - redis
    networks:
      - go-api-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/go-api/external/service/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  mysql:
    image: mysql:8.0
    container_name: go-api-mysql-prod
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: go_api_prod
      MYSQL_USER: go_api_user
      MYSQL_PASSWORD: ${DB_PASSWORD}
      TZ: Asia/Shanghai
    volumes:
      - mysql_data:/var/lib/mysql
      - ./bin/data/sql:/docker-entrypoint-initdb.d:ro
    ports:
      - "3306:3306"
    networks:
      - go-api-network
    command: --default-authentication-plugin=mysql_native_password
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p${DB_ROOT_PASSWORD}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    container_name: go-api-redis-prod
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - go-api-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    container_name: go-api-nginx-prod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./nginx/logs:/var/log/nginx
    depends_on:
      - go-api
    networks:
      - go-api-network

volumes:
  mysql_data:
  redis_data:

networks:
  go-api-network:
    driver: bridge
```

#### 2. Nginx Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream go-api {
        server go-api:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

    server {
        listen 80;
        server_name your-domain.com;

        # Redirect HTTP to HTTPS
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        # SSL Configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

        # API routes
        location /go-api/ {
            limit_req zone=api burst=20 nodelay;
            
            proxy_pass http://go-api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # Health check
        location /health {
            proxy_pass http://go-api/go-api/external/service/ping;
        }

        # Static files (if any)
        location /static/ {
            alias /var/www/static/;
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
}
```

#### 3. Deploy with Docker Compose

```bash
# Create environment file
cat > .env << EOF
DB_ROOT_PASSWORD=secure_root_password
DB_PASSWORD=secure_database_password
REDIS_PASSWORD=secure_redis_password
EOF

# Build and deploy
docker-compose -f docker-compose.prod.yml up -d

# Check services
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f go-api
```

### Kubernetes Deployment

#### 1. Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: go-api

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-api-config
  namespace: go-api
data:
  prod.json: |
    {
      "system": {
        "name": "go-api",
        "run_mode": "release",
        "http_port": ":8080"
      }
    }
```

#### 2. Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: go-api-secrets
  namespace: go-api
type: Opaque
data:
  jwt-secret: base64_encoded_jwt_secret
  db-password: base64_encoded_db_password
  redis-password: base64_encoded_redis_password
```

#### 3. Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-api
  namespace: go-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-api
  template:
    metadata:
      labels:
        app: go-api
    spec:
      containers:
      - name: go-api
        image: your-registry/go-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: RUN_ENV
          value: "prod"
        - name: APP_NAME
          value: "go-api"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: go-api-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /bin/configs
        - name: logs
          mountPath: /bin/logs
        livenessProbe:
          httpGet:
            path: /go-api/external/service/ping
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /go-api/external/service/ping
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: go-api-config
      - name: logs
        emptyDir: {}
```

#### 4. Service and Ingress

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-api-service
  namespace: go-api
spec:
  selector:
    app: go-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-api-ingress
  namespace: go-api
  annotations:
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: go-api-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /go-api
        pathType: Prefix
        backend:
          service:
            name: go-api-service
            port:
              number: 80
```

### Monitoring and Logging

#### 1. Prometheus Metrics

Add metrics endpoint to your application:

```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: go-api-metrics
  namespace: go-api
spec:
  selector:
    matchLabels:
      app: go-api
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

#### 2. Log Aggregation

Use ELK Stack or similar:

```yaml
# k8s/logstash-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: logstash-config
data:
  logstash.conf: |
    input {
      beats {
        port => 5044
      }
    }
    filter {
      if [kubernetes][container][name] == "go-api" {
        json {
          source => "message"
        }
      }
    }
    output {
      elasticsearch {
        hosts => ["elasticsearch:9200"]
        index => "go-api-logs-%{+YYYY.MM.dd}"
      }
    }
```

### Security Considerations

#### 1. Network Security

- Use private networks for database and Redis
- Enable firewall rules
- Use VPN for administrative access

#### 2. Application Security

- Regular security updates
- Environment variable encryption
- Secret management
- Rate limiting
- Input validation

#### 3. SSL/TLS Configuration

```bash
# Generate SSL certificate (Let's Encrypt)
certbot certonly --webroot -w /var/www/html -d your-domain.com

# Or use existing certificates
cp your-cert.pem nginx/ssl/cert.pem
cp your-key.pem nginx/ssl/key.pem
```

### Backup and Recovery

#### 1. Database Backup

```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker exec go-api-mysql-prod mysqldump -u root -p${DB_ROOT_PASSWORD} go_api_prod > backup_${DATE}.sql
gzip backup_${DATE}.sql
aws s3 cp backup_${DATE}.sql.gz s3://your-backup-bucket/
EOF

# Schedule with cron
echo "0 2 * * * /path/to/backup.sh" | crontab -
```

#### 2. Application Data Backup

```bash
# Backup configuration and data
tar -czf go-api-backup-$(date +%Y%m%d).tar.gz bin/configs bin/data

# Upload to cloud storage
aws s3 cp go-api-backup-$(date +%Y%m%d).tar.gz s3://your-backup-bucket/
```

### Performance Optimization

#### 1. Database Optimization

```sql
-- Add indexes for better performance
ALTER TABLE auth_app ADD INDEX idx_app_id (app_id);
ALTER TABLE auth_app ADD INDEX idx_status (status);
```

#### 2. Caching Strategy

- Redis for session data
- Application-level caching
- CDN for static assets

#### 3. Load Balancing

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: go-api-hpa
  namespace: go-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: go-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

---

## 中文

### 前提条件

- Docker 20.10+ 和 Docker Compose 2.0+
- Linux/macOS/Windows 服务器
- 最少 1GB RAM，1 CPU 核心
- 数据库和 Redis 的网络访问

### 环境设置

#### 1. 生产配置

创建生产配置文件：

```bash
# 创建生产配置
cp bin/configs/local.json.default bin/configs/prod.json
```

编辑 `bin/configs/prod.json`：

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "release",
    "http_port": ":8080",
    "read_timeout": 60,
    "write_timeout": 60,
    "jwt_secret": "你的生产环境密钥-请更改此项",
    "default_lang": "zh-CN"
  },
  "log": {
    "driver": "file",
    "level": "info",
    "path": "/bin/logs/"
  },
  "databases": [
    {
      "enable": true,
      "db_type": "mysql",
      "db_name": "go_api_prod",
      "db_host": "mysql:3306",
      "db_username": "go_api_user",
      "db_password": "安全密码在这里",
      "db_max_idle_conn": 10,
      "db_max_open_conn": 100,
      "db_max_lifetime": 3
    }
  ],
  "redis": [
    {
      "enable": true,
      "name": "go-api",
      "host": "redis:6379",
      "auth": "redis密码在这里",
      "max_idle": 30,
      "max_active": 100,
      "idle_timeout": 30,
      "prefix": "go-api-prod",
      "db": 0
    }
  ]
}
```

### Docker 部署

#### 1. Docker Compose（推荐）

创建 `docker-compose.prod.yml`：

```yaml
version: '3.8'

services:
  go-api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: go-api-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./bin/configs:/bin/configs:ro
      - ./bin/logs:/bin/logs
      - ./bin/data:/bin/data:ro
    environment:
      - RUN_ENV=prod
      - APP_NAME=go-api
      - TZ=Asia/Shanghai
    depends_on:
      - mysql
      - redis
    networks:
      - go-api-network

  mysql:
    image: mysql:8.0
    container_name: go-api-mysql-prod
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: go_api_prod
      MYSQL_USER: go_api_user
      MYSQL_PASSWORD: ${DB_PASSWORD}
      TZ: Asia/Shanghai
    volumes:
      - mysql_data:/var/lib/mysql
      - ./bin/data/sql:/docker-entrypoint-initdb.d:ro
    ports:
      - "3306:3306"
    networks:
      - go-api-network

  redis:
    image: redis:7-alpine
    container_name: go-api-redis-prod
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - go-api-network

volumes:
  mysql_data:
  redis_data:

networks:
  go-api-network:
    driver: bridge
```

#### 2. 使用 Docker Compose 部署

```bash
# 创建环境文件
cat > .env << EOF
DB_ROOT_PASSWORD=安全的root密码
DB_PASSWORD=安全的数据库密码
REDIS_PASSWORD=安全的redis密码
EOF

# 构建和部署
docker-compose -f docker-compose.prod.yml up -d

# 检查服务
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f go-api
```

### 监控和日志

#### 1. 数据库备份

```bash
# 创建备份脚本
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker exec go-api-mysql-prod mysqldump -u root -p${DB_ROOT_PASSWORD} go_api_prod > backup_${DATE}.sql
gzip backup_${DATE}.sql
# 上传到云存储
aws s3 cp backup_${DATE}.sql.gz s3://your-backup-bucket/
EOF

# 使用 cron 调度
echo "0 2 * * * /path/to/backup.sh" | crontab -
```

#### 2. 应用数据备份

```bash
# 备份配置和数据
tar -czf go-api-backup-$(date +%Y%m%d).tar.gz bin/configs bin/data

# 上传到云存储
aws s3 cp go-api-backup-$(date +%Y%m%d).tar.gz s3://your-backup-bucket/
```

### 性能优化

#### 1. 数据库优化

```sql
-- 添加索引以提高性能
ALTER TABLE auth_app ADD INDEX idx_app_id (app_id);
ALTER TABLE auth_app ADD INDEX idx_status (status);
```

#### 2. 缓存策略

- Redis 用于会话数据
- 应用级缓存
- CDN 用于静态资源

### 安全考虑

#### 1. 网络安全

- 为数据库和 Redis 使用私有网络
- 启用防火墙规则
- 使用 VPN 进行管理访问

#### 2. 应用安全

- 定期安全更新
- 环境变量加密
- 密钥管理
- 速率限制
- 输入验证