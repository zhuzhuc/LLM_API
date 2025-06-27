#!/bin/bash

# 切换回 SQLite 数据库的脚本

echo "🔄 切换回 SQLite 数据库..."

# 备份当前的 .env 文件
if [ -f .env ]; then
    cp .env .env.backup
    echo "✅ 已备份当前配置到 .env.backup"
fi

# 更新 .env 文件
if [ -f .env ]; then
    # 注释掉 MySQL 配置
    sed -i.bak 's/^DATABASE_URL=mysql/#DATABASE_URL=mysql/' .env
    
    # 启用 SQLite 配置
    if grep -q "#DATABASE_URL=sqlite3" .env; then
        sed -i.bak 's/^#DATABASE_URL=sqlite3/DATABASE_URL=sqlite3/' .env
    else
        # 如果没有找到注释的 SQLite 配置，添加一个
        echo "DATABASE_URL=sqlite3://./llm.db" >> .env
    fi
    
    rm -f .env.bak
    echo "✅ 已更新 .env 文件使用 SQLite"
else
    echo "❌ .env 文件不存在"
    exit 1
fi

echo ""
echo "🎉 SQLite 配置完成！"
echo "📝 请重新启动服务以应用新的数据库配置:"
echo "   ./stop-llm-service.sh"
echo "   ./start-llm-service.sh"
echo ""
echo "🔄 如需切换到 MySQL，请运行:"
echo "   ./switch-to-mysql.sh"
