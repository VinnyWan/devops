package api

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
)

func getUserInfo(c *gin.Context) (uint, string) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	uID, _ := userID.(uint)
	uName, _ := username.(string)
	return uID, uName
}

// FileBrowse lists files and directories at a given path on a host.
func FileBrowse(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.BrowseRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, _ := getUserInfo(c)
	resp, err := svc.Browse(tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "浏览文件失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    resp,
	})
}

// FileDownload downloads a file from a host and streams it to the client.
func FileDownload(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.DownloadRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	var buf bytes.Buffer
	n, host, svcErr := svc.Download(tenantID, userID, req.HostID, req.FilePath, &buf)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "download", req.FilePath, 0, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "下载文件失败", "error": svcErr.Error()})
		return
	}

	fileName := filepath.Base(req.FilePath)
	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, host.Ip, host.Hostname, "download", req.FilePath, n, "success", "")

	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
}

// FileUpload uploads a file to a host.
func FileUpload(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	hostIDStr := c.Param("hostId")
	hostID, err := strconv.ParseUint(hostIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的主机 ID"})
		return
	}

	remotePath := c.PostForm("path")
	if remotePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少目标路径"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取上传文件失败", "error": err.Error()})
		return
	}
	defer file.Close()

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	uploadReq := service.UploadRequest{
		HostID:     uint(hostID),
		RemotePath: remotePath,
	}
	svcErr := svc.Upload(tenantID, userID, uploadReq, file, header.Size)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, uint(hostID), "", "", "upload", remotePath, header.Size, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传文件失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, uint(hostID), "", "", "upload", remotePath, header.Size, "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
	})
}

// FileDelete deletes a file or directory on a host.
func FileDelete(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	svcErr := svc.Delete(tenantID, userID, req.HostID, req.FilePath)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "delete", req.FilePath, 0, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "delete", req.FilePath, 0, "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// FileRename renames a file or directory on a host.
func FileRename(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	svcErr := svc.Rename(tenantID, userID, req)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "rename", req.OldPath+" -> "+req.NewPath, 0, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重命名失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "rename", req.OldPath+" -> "+req.NewPath, 0, "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "重命名成功",
	})
}

// FileMkdir creates a directory on a host.
func FileMkdir(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.MkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	svcErr := svc.Mkdir(tenantID, userID, req.HostID, req.Path)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "mkdir", req.Path, 0, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建目录失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "mkdir", req.Path, 0, "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建目录成功",
	})
}

// FileChmod changes file permissions on a host.
func FileChmod(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.ChmodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	svcErr := svc.Chmod(tenantID, userID, req)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "chmod", req.Path, 0, "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "修改权限失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "chmod", req.Path, 0, "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "修改权限成功",
	})
}

// FilePreview reads file content (up to 1 MB) for preview.
func FilePreview(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.ReadFileRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, _ := getUserInfo(c)
	content, svcErr := svc.ReadFile(tenantID, userID, req.HostID, req.FilePath)

	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "预览文件失败", "error": svcErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"content": content,
			"path":    req.FilePath,
		},
	})
}

// FileEdit writes content to a file on a host.
func FileEdit(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.WriteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)
	clientIP := c.ClientIP()

	svcErr := svc.WriteFile(tenantID, userID, req)

	if svcErr != nil {
		svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "edit", req.FilePath, int64(len(req.Content)), "failure", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "编辑文件失败", "error": svcErr.Error()})
		return
	}

	svc.RecordAudit(tenantID, userID, username, clientIP, req.HostID, "", "", "edit", req.FilePath, int64(len(req.Content)), "success", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "编辑成功",
	})
}

// FileDistribute distributes a file to multiple hosts.
func FileDistribute(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	file, _, formErr := c.Request.FormFile("file")
	if formErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取上传文件失败", "error": formErr.Error()})
		return
	}
	defer file.Close()

	content, readErr := io.ReadAll(file)
	if readErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "读取文件内容失败", "error": readErr.Error()})
		return
	}

	remotePath := c.PostForm("path")
	if remotePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少目标路径"})
		return
	}

	hostIdsStr := c.PostForm("hostIds")
	if hostIdsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少目标主机"})
		return
	}

	var hostIDs []uint
	for _, idStr := range strings.Split(hostIdsStr, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, parseErr := strconv.ParseUint(idStr, 10, 64)
		if parseErr != nil {
			continue
		}
		hostIDs = append(hostIDs, uint(id))
	}
	if len(hostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无有效的主机 ID"})
		return
	}

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	userID, username := getUserInfo(c)

	req := service.DistributeRequest{
		HostIDs:    hostIDs,
		RemotePath: remotePath,
	}
	results := svc.Distribute(tenantID, userID, req, content)

	// Record audit for each host
	clientIP := c.ClientIP()
	for _, r := range results {
		result := "success"
		errMsg := ""
		if !r.Success {
			result = "failure"
			errMsg = r.Error
		}
		svc.RecordAudit(tenantID, userID, username, clientIP, r.HostID, r.HostIP, r.HostName, "distribute", remotePath, int64(len(content)), result, errMsg)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分发完成",
		"data":    results,
	})
}

// FileAuditList returns paginated file operation audit logs.
func FileAuditList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	svc := getFileService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	req := service.ListAuditRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		OpType:   c.Query("opType"),
		Username: c.Query("username"),
		HostIP:   c.Query("hostIp"),
		StartAt:  c.Query("startAt"),
		EndAt:    c.Query("endAt"),
	}

	logs, total, svcErr := svc.ListAudit(tenantID, req)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取审计日志失败", "error": svcErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
