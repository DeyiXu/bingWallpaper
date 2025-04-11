#!/bin/bash

# https://github.com/DeyiXu/bingWallpaper
# 脚本功能：使用bingWallpaper工具自动下载必应壁纸并设置为系统壁纸
# 用法：./auto-set-wallpaper.sh

# 设置壁纸存储目录
WALLPAPER_DIR="$HOME/Pictures/bing_wallpapers"

# 创建壁纸存储目录（如果不存在）
mkdir -p "$WALLPAPER_DIR"

# 壁纸文件路径 - 使用绝对路径避免路径嵌套
WALLPAPER_PATH="$WALLPAPER_DIR/desktop_background.jpg"

echo "正在使用bingWallpaper工具下载必应壁纸..."

# 切换到工作目录，避免bingWallpaper相对路径问题
cd "$WALLPAPER_DIR" || exit 1
echo "工作目录: $(pwd)"

# 使用-name参数时提供纯文件名，而非完整路径
WALLPAPER_FILENAME="desktop_background.jpg"
bingWallpaper -last -dir "$WALLPAPER_DIR" -name "$WALLPAPER_FILENAME" -overwrite

# 确保检查正确位置的文件
WALLPAPER_PATH="$WALLPAPER_DIR/$WALLPAPER_FILENAME"

# 检查壁纸是否成功下载
if [ ! -f "$WALLPAPER_PATH" ]; then
    echo "错误：未能下载壁纸到 $WALLPAPER_PATH"
    # 尝试查找任何下载的图片
    echo "尝试查找任何已下载的壁纸..."
    FOUND_WALLPAPER=$(find "$WALLPAPER_DIR" -maxdepth 2 -type f -name "*.jpg" | head -n 1)
    if [ -n "$FOUND_WALLPAPER" ]; then
        echo "找到壁纸: $FOUND_WALLPAPER"
        # 将找到的壁纸复制到预期位置
        cp "$FOUND_WALLPAPER" "$WALLPAPER_PATH"
        echo "已复制壁纸到: $WALLPAPER_PATH"
    else
        echo "无法找到任何壁纸文件"
        exit 1
    fi
fi

echo "壁纸成功下载到: $WALLPAPER_PATH"

# 自动设置系统壁纸的函数
set_wallpaper() {
    local wallpaper="$1"
    echo "正在尝试设置壁纸: $wallpaper"
    
    # 确认文件存在
    if [ ! -f "$wallpaper" ]; then
        echo "错误: 壁纸文件不存在: $wallpaper"
        return 1
    fi
    
    # 完整路径的壁纸文件
    WALLPAPER_FULLPATH=$(realpath "$wallpaper")
    echo "壁纸完整路径: $WALLPAPER_FULLPATH"
    
    # 扩展桌面环境检测
    # 检测桌面环境
    if [[ "$XDG_CURRENT_DESKTOP" == *"GNOME"* ]] || [[ "$DESKTOP_SESSION" == *"gnome"* ]]; then
        echo "检测到GNOME桌面环境"
        gsettings set org.gnome.desktop.background picture-uri "file:$WALLPAPER_FULLPATH"
        gsettings set org.gnome.desktop.background picture-uri-dark "file:$WALLPAPER_FULLPATH"
        echo "已设置GNOME壁纸"
    elif [[ "$XDG_CURRENT_DESKTOP" == *"KDE"* ]] || [[ "$DESKTOP_SESSION" == *"plasma"* ]]; then
        echo "检测到KDE桌面环境"
        dbus-send --session --dest=org.kde.plasmashell --type=method_call /PlasmaShell org.kde.PlasmaShell.evaluateScript "string:var Desktops = desktops();for (i=0;i<Desktops.length;i++){d = Desktops[i];d.wallpaperPlugin = 'org.kde.image';d.currentConfigGroup = Array('Wallpaper','org.kde.image','General');d.writeConfig('Image','file:$WALLPAPER_FULLPATH');}"
        echo "已设置KDE壁纸"
    elif [[ "$XDG_CURRENT_DESKTOP" == *"XFCE"* ]] || [[ "$DESKTOP_SESSION" == *"xfce"* ]]; then
        echo "检测到Xfce桌面环境"
        xfconf-query -c xfce4-desktop -p /backdrop/screen0/monitor0/workspace0/last-image -s "$WALLPAPER_FULLPATH"
        echo "已设置Xfce壁纸"
    elif [[ "$XDG_CURRENT_DESKTOP" == *"MATE"* ]]; then
        echo "检测到MATE桌面环境"
        gsettings set org.mate.background picture-filename "$WALLPAPER_FULLPATH"
        echo "已设置MATE壁纸"
    elif [[ "$XDG_CURRENT_DESKTOP" == *"Cinnamon"* ]]; then
        echo "检测到Cinnamon桌面环境"
        gsettings set org.cinnamon.desktop.background picture-uri "file:$WALLPAPER_FULLPATH"
        echo "已设置Cinnamon壁纸"
    elif command -v feh >/dev/null 2>&1; then
        echo "使用feh设置壁纸"
        feh --bg-fill "$WALLPAPER_FULLPATH"
        echo "已使用feh设置壁纸"
    elif command -v nitrogen >/dev/null 2>&1; then
        echo "使用nitrogen设置壁纸"
        nitrogen --set-scaled "$WALLPAPER_FULLPATH"
        echo "已使用nitrogen设置壁纸"
    # 尝试pcmanfm (常用于LXDE)
    elif command -v pcmanfm >/dev/null 2>&1; then
        echo "尝试使用pcmanfm设置壁纸"
        pcmanfm --set-wallpaper="$WALLPAPER_FULLPATH"
        echo "已尝试使用pcmanfm设置壁纸"
    # 尝试使用hsetroot
    elif command -v hsetroot >/dev/null 2>&1; then
        echo "使用hsetroot设置壁纸"
        hsetroot -fill "$WALLPAPER_FULLPATH"
        echo "已使用hsetroot设置壁纸"
    # 使用标准的xwallpaper
    elif command -v xwallpaper >/dev/null 2>&1; then
        echo "使用xwallpaper设置壁纸"
        xwallpaper --zoom "$WALLPAPER_FULLPATH"
        echo "已使用xwallpaper设置壁纸" 
    else
        echo "无法检测到支持的桌面环境或壁纸设置工具"
        echo "如果你想自动设置壁纸，请安装 feh、nitrogen 或 pcmanfm 等工具"
        return 1
    fi
    
    return 0
}

# 调用设置壁纸函数
echo "正在尝试自动设置系统壁纸..."
if [ -f "$WALLPAPER_PATH" ]; then
    if set_wallpaper "$WALLPAPER_PATH"; then
        echo "壁纸设置成功!"
    else
        echo "壁纸设置失败，请手动设置或安装支持的壁纸设置工具"
    fi
else
    echo "错误：找不到壁纸文件，无法设置壁纸"
fi
