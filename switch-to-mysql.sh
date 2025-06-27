#!/bin/bash

# 切换到 MySQL 数据库的脚本

echo "🔄 切换到 MySQL 数据库..."

# 检查 MySQL 是否安装
if ! command -v mysql &> /dev/null; then
    echo "❌ MySQL 未安装，请先安装 MySQL"
    echo "macOS: brew install mysql"
    echo "Ubuntu: sudo apt-get install mysql-server"
    echo "CentOS: sudo yum install mysql-server"
    exit 1
fi

# 检查 MySQL 服务是否运行
if ! pgrep -x "mysqld" > /dev/null; then
    echo "⚠️  MySQL 服务未运行，正在启动..."
    if command -v brew &> /dev/null; then
        # macOS with Homebrew
        brew services start mysql
    elif command -v systemctl &> /dev/null; then
        # Linux with systemd
        sudo systemctl start mysql
    else
        echo "❌ 无法启动 MySQL 服务，请手动启动"
        exit 1
    fi
fi

# 提示用户设置数据库
echo "📋 请按照以下步骤设置 MySQL 数据库："
echo ""
echo "1. 连接到 MySQL (使用 root 用户):"
echo "   mysql -u root -p"
echo ""
echo "2. 执行数据库设置脚本:"
echo "   source setup-mysql.sql"
echo ""
echo "3. 或者手动执行以下命令:"
echo "   CREATE DATABASE llm_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
echo "   CREATE USER '123456'@'localhost' IDENTIFIED BY 'llm_password';"
echo "   GRANT ALL PRIVILEGES ON llm_db.* TO '123456'@'localhost';"
echo "   FLUSH PRIVILEGES;"
echo ""

# 询问是否已经设置好数据库
read -p "是否已经设置好 MySQL 数据库? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # 备份当前的 .env 文件
    if [ -f .env ]; then
        cp .env .env.backup
        echo "✅ 已备份当前配置到 .env.backup"
    fi
    
    # 提示用户输入数据库连接信息
    echo ""
    echo "请输入 MySQL 连接信息:"
    read -p "数据库主机 (默认: localhost): " DB_HOST
    DB_HOST=${DB_HOST:-localhost}
    
    read -p "数据库端口 (默认: 3306): " DB_PORT
    DB_PORT=${DB_PORT:-3306}
    
    read -p "数据库名称 (默认: llm_db): " DB_NAME
    DB_NAME=${DB_NAME:-llm_db}
    
    read -p "用户名 (默认: llm_user): " DB_USER
    DB_USER=${DB_USER:-llm_user}
    
    read -s -p "密码: " DB_PASSWORD
    echo
    
    # 构建 MySQL 连接字符串
    MYSQL_URL="mysql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    
    # 更新 .env 文件
    if [ -f .env ]; then
        # 注释掉 SQLite 配置
        sed -i.bak 's/^DATABASE_URL=sqlite3/#DATABASE_URL=sqlite3/' .env
        
        # 添加或更新 MySQL 配置
        if grep -q "DATABASE_URL=mysql" .env; then
            sed -i.bak "s|^.*DATABASE_URL=mysql.*|DATABASE_URL=${MYSQL_URL}|" .env
        else
            echo "DATABASE_URL=${MYSQL_URL}" >> .env
        fi
        
        rm -f .env.bak
        echo "✅ 已更新 .env 文件使用 MySQL"
    else
        echo "❌ .env 文件不存在"
        exit 1
    fi
    
    echo ""
    echo "🎉 MySQL 配置完成！"
    echo "📝 请重新启动服务以应用新的数据库配置:"
    echo "   ./stop-llm-service.sh"
    echo "   ./start-llm-service.sh"
    echo ""
    echo "🔄 如需切换回 SQLite，请运行:"
    echo "   ./switch-to-sqlite.sh"
else
    echo ""
    echo "请先设置 MySQL 数据库，然后重新运行此脚本"
fi
