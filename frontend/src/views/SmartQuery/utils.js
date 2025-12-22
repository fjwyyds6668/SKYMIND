// 获取删除确认内容
export const getDeleteConfirmContent = (element, type = 'topic') => {
  if (type === 'assistant') {
    return `确定要删除助手"${element.name}"吗？删除后将无法恢复，且会删除该助手下的所有话题、对话和消息。`;
  } else {
    return `确定要删除话题"${element.name}"吗？删除后将无法恢复，同时会删除该话题下的所有对话和消息。`;
  }
};

// 格式化话题时间
export const formatTopicTime = (timeString) => {
  if (!timeString) return "";
  const date = new Date(timeString);
  const now = new Date();
  const diffMs = now - date;
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) {
    // 今天
    return date.toLocaleTimeString('zh-CN', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  } else if (diffDays === 1) {
    // 昨天
    return '昨天 ' + date.toLocaleTimeString('zh-CN', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  } else if (diffDays < 7) {
    // 本周
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    return weekdays[date.getDay()] + ' ' + date.toLocaleTimeString('zh-CN', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  } else {
    // 更早
    return date.toLocaleDateString('zh-CN', { 
      month: '2-digit', 
      day: '2-digit' 
    });
  }
};

// 格式化系统时间
export const formatSystemTime = () => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth() + 1;
  const date = now.getDate();
  
  // 计算周数
  const startDate = new Date(now.getFullYear(), 0, 1);
  const days = Math.floor((now - startDate) / (24 * 60 * 60 * 1000));
  const weekNumber = Math.ceil((days + startDate.getDay() + 1) / 7);
  
  // 星期
  const weekdays = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六'];
  const weekday = weekdays[now.getDay()];
  
  // 时间
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  const seconds = String(now.getSeconds()).padStart(2, '0');
  
  return `${year}年${month}月${date}日 第${weekNumber}周 ${weekday} ${hours}:${minutes}:${seconds}`;
};

// 格式化日期时间
export const formatDateTime = (datetime) => {
  if (!datetime) return new Date().toLocaleString();
  return datetime instanceof Date ? datetime.toLocaleString() : new Date(datetime).toLocaleString();
};

// 生成UUID
export const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
};

// 计算文件MD5（使用简单的哈希函数替代MD5）
export const calculateMD5 = async (file) => {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = async (e) => {
      const buffer = e.target.result;
      // 使用SHA-256替代MD5，因为Web Crypto API支持SHA-256
      const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
      resolve(hashHex);
    };
    reader.readAsArrayBuffer(file);
  });
};

// 压缩图片
export const compressImage = (file, maxWidth = 1920, maxHeight = 1080, quality = 0.8) => {
  return new Promise((resolve) => {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    const img = new Image();
    
    img.onload = () => {
      // 计算压缩后的尺寸
      let { width, height } = img;
      
      // 如果图片尺寸大于最大尺寸，按比例缩放
      if (width > maxWidth || height > maxHeight) {
        const ratio = Math.min(maxWidth / width, maxHeight / height);
        width *= ratio;
        height *= ratio;
      }
      
      canvas.width = width;
      canvas.height = height;
      
      // 绘制压缩后的图片
      ctx.drawImage(img, 0, 0, width, height);
      
      // 转换为Blob
      canvas.toBlob((blob) => {
        resolve(blob);
      }, 'image/jpeg', quality);
    };
    
    img.src = URL.createObjectURL(file);
  });
};

// 处理图片文件压缩
export const processImageFile = async (file) => {
  try {
    // 检查文件大小，如果小于10MB，可能不需要压缩
    const maxSize = 10 * 1024 * 1024; // 10MB
    let processedFile = file;
    
    if (file.size > maxSize) {
      // 压缩图片
      const compressedBlob = await compressImage(file, 1920, 1080, 0.7);
      processedFile = new File([compressedBlob], file.name, {
        type: 'image/jpeg',
        lastModified: Date.now()
      });
    }
    
    // 生成文件信息
    const fileName = file.name.replace(/\.[^/.]+$/, ""); // 去掉后缀
    const fileSuffix = file.name.split('.').pop().toLowerCase();
    const uuid = generateUUID();
    const md5 = await calculateMD5(processedFile);
    
    return {
      originalFile: file,
      processedFile: processedFile,
      fileName: fileName,
      fileSuffix: fileSuffix,
      uuid: uuid,
      md5: md5,
      originalPath: file.path || '',
      size: processedFile.size
    };
  } catch (error) {
    console.error('图片处理失败:', error);
    throw error;
  }
};

// 处理非图片文件
export const processOtherFile = async (file) => {
  try {
    const fileName = file.name.replace(/\.[^/.]+$/, ""); // 去掉后缀
    const fileSuffix = file.name.split('.').pop().toLowerCase();
    const uuid = generateUUID();
    const md5 = await calculateMD5(file);
    
    return {
      originalFile: file,
      processedFile: file,
      fileName: fileName,
      fileSuffix: fileSuffix,
      uuid: uuid,
      md5: md5,
      originalPath: file.path || '',
      size: file.size
    };
  } catch (error) {
    console.error('文件处理失败:', error);
    throw error;
  }
};

// 判断是否为图片文件
export const isImageFile = (fileName) => {
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'];
  const extension = fileName.split('.').pop().toLowerCase();
  return imageExtensions.includes(extension);
};

// 统一文件处理入口
export const processFile = async (file) => {
  if (isImageFile(file.name)) {
    return await processImageFile(file);
  } else {
    return await processOtherFile(file);
  }
};

// 格式化文件大小显示
export const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
