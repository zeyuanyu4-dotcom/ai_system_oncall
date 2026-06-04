package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceService *service.ServiceService
}

func NewServiceHandler(serviceService *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

// CreateService creates a new service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	req.ProjectID = projectID
	serviceInfo, err := h.serviceService.CreateService(&req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, serviceInfo)
}

// GetService gets a service by ID
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	serviceInfo, err := h.serviceService.GetService(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	response.Success(c, serviceInfo)
}

// ListServices lists services
func (h *ServiceHandler) ListServices(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	var req dto.ServiceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	req.ProjectID = projectID
	result, err := h.serviceService.ListServices(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ListAllServices lists all services (not project-specific)
func (h *ServiceHandler) ListAllServices(c *gin.Context) {
	var req dto.ServiceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.serviceService.ListAllServices(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateService updates a service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	serviceInfo, err := h.serviceService.UpdateService(id, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, serviceInfo)
}

// DeleteService deletes a service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	if err := h.serviceService.DeleteService(id); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

type ServiceAPIHandler struct {
	apiService *service.ServiceAPIService
}

func NewServiceAPIHandler(apiService *service.ServiceAPIService) *ServiceAPIHandler {
	return &ServiceAPIHandler{apiService: apiService}
}

// CreateAPI creates a new service API
func (h *ServiceAPIHandler) CreateAPI(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	var req dto.CreateServiceAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	apiInfo, err := h.apiService.CreateAPI(serviceID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, apiInfo)
}

// ListAPIs lists APIs of a service
func (h *ServiceAPIHandler) ListAPIs(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	apis, err := h.apiService.ListAPIs(serviceID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, apis)
}

// UpdateAPI updates a service API
func (h *ServiceAPIHandler) UpdateAPI(c *gin.Context) {
	apiID, err := strconv.ParseUint(c.Param("api_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	var req dto.UpdateServiceAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	apiInfo, err := h.apiService.UpdateAPI(apiID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, apiInfo)
}

// DeleteAPI deletes a service API
func (h *ServiceAPIHandler) DeleteAPI(c *gin.Context) {
	apiID, err := strconv.ParseUint(c.Param("api_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	if err := h.apiService.DeleteAPI(apiID); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

type ServiceDependencyHandler struct {
	depService *service.ServiceDependencyService
}

func NewServiceDependencyHandler(depService *service.ServiceDependencyService) *ServiceDependencyHandler {
	return &ServiceDependencyHandler{depService: depService}
}

// CreateDependency creates a new service dependency
func (h *ServiceDependencyHandler) CreateDependency(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	var req dto.CreateServiceDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	depInfo, err := h.depService.CreateDependency(serviceID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, depInfo)
}

// ListDependencies lists dependencies of a service
func (h *ServiceDependencyHandler) ListDependencies(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	deps, err := h.depService.ListDependencies(serviceID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, deps)
}

// DeleteDependency deletes a service dependency
func (h *ServiceDependencyHandler) DeleteDependency(c *gin.Context) {
	depID, err := strconv.ParseUint(c.Param("dependency_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的依赖ID")
		return
	}

	if err := h.depService.DeleteDependency(depID); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}
