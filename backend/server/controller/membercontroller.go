package controller

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/carsonalh/churchmanagerbackend/server/domain"
	"github.com/carsonalh/churchmanagerbackend/server/store"
	"github.com/gin-gonic/gin"
)

type MemberController struct {
	store           *store.MemberStore
	defaultPageSize uint
	maxPageSize     uint
}

type MemberControllerConfig struct {
	DefaultPageSize uint
	MaxPageSize     uint
}

func SetupMemberController(router *gin.RouterGroup, store *store.MemberStore, config *MemberControllerConfig) *MemberController {
	controller := &MemberController{
		store:           store,
		maxPageSize:     config.MaxPageSize,
		defaultPageSize: config.DefaultPageSize,
	}

	router.GET("", controller.getMembers)
	router.POST("", controller.postMember)
	router.GET(":id", controller.getMember)
	router.PUT(":id", controller.putMember)
	router.DELETE(":id", controller.deleteMember)

	return controller
}

// getMembers godoc
// @Summary      Get index of members.
// @Description  Invalid query parameters are coerced to their default values.
// @Param        pageSize query int false "The size of the returned page. Maximum value is 500."
// @Param        page     query int false "The page index (zero-based) to get. Pages that are out of range return emtpy lists."
// @Accept       json
// @Produce      json
// @Success      200 {array} domain.MemberResponseDTO
// @Router       /members [get]
func (controller *MemberController) getMembers(c *gin.Context) {
	var members []domain.Member
	var err error

	pageSize64, err := strconv.ParseUint(c.Query("pageSize"), 10, 32)
	var pageSize uint
	if err != nil {
		pageSize = controller.defaultPageSize
	} else {
		pageSize = uint(pageSize64)
	}
	pageSize = min(pageSize, controller.maxPageSize)

	page64, err := strconv.ParseUint(c.Query("page"), 10, 32)
	var page uint
	if err != nil {
		page = 0
	} else {
		page = uint(page64)
	}

	if members, err = controller.store.GetPage(pageSize, page); err != nil {
		log.Printf("GET /members : error getting members from database: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	responseDTOs := make([]domain.MemberResponseDTO, 0)

	for _, member := range members {
		responseDTOs = append(responseDTOs, *member.ToResponseDTO())
	}

	c.JSON(http.StatusOK, responseDTOs)
}

// getMember godoc
// @Summary      Get a member
// @Param        id path int true "The id of the member to get"
// @Accept       json
// @Produce      json
// @Success      200 {object} domain.MemberResponseDTO
// @Failure      400 The id could not be parsed into an integer of appropriate size
// @Router       /members/{id} [get]
func (controller *MemberController) getMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid parameter id \"%s\"\n", c.Param("id"))
		return
	}

	member, err := controller.store.FindById(id)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if member == nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, member.ToResponseDTO())
	}
}

// postMember godoc
// @Summary      Add a member
// @Param        request body domain.MemberUpdateDTO true "Member to add"
// @Accept       json
// @Produce      json
// @Success      201 {object} domain.MemberResponseDTO
// @Failure      400 Invalid input data
// @Router       /members [post]
func (controller *MemberController) postMember(c *gin.Context) {
	// Create and update are the same DTO
	var createDto domain.MemberUpdateDTO

	if err := c.BindJSON(&createDto); err != nil {
		c.String(http.StatusBadRequest, err.Error()+"\n")
		return
	}

	if errs := createDto.Validate(); len(errs) > 0 {
		builder := strings.Builder{}
		builder.WriteString("Failed to validate create object with the following errors:\n")
		for _, err := range errs {
			builder.WriteString(err.Error())
			builder.WriteString("\n")
		}
		c.String(http.StatusBadRequest, builder.String())
		return
	}

	member, err := controller.store.Create(&createDto)
	if err != nil {
		log.Printf("failed to create member: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	idString := strconv.FormatUint(member.Id(), 10)
	c.Header("Location", c.Request.URL.Path+"/"+idString)
	c.JSON(http.StatusCreated, member.ToResponseDTO())
}

// deleteMember godoc
// @Summary      Delete a member
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Member ID"
// @Success      200
// @Failure      404 No member with the given id could be found to delete
// @Router       /members/{id} [delete]
func (controller *MemberController) deleteMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id \"%s\"\n", c.Param("id"))
		return
	}

	deleted, err := controller.store.DeleteById(id)
	if err != nil {
		log.Printf("error deleting member by id: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !deleted {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

type putMember struct {
	Id uint64 `uri:"id" binding:"required"`
	domain.MemberUpdateDTO
}

// putMember godoc
// @Summary      Update a member
// @Param        request body domain.MemberUpdateDTO true "New data for the member. This operation replaces the member entirely."
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Member ID"
// @Success      200 {object} domain.MemberResponseDTO
// @Router       /members/{id} [put]
func (c *MemberController) putMember(ctx *gin.Context) {
	var request putMember

	if err := ctx.BindUri(&request); err != nil {
		ctx.String(http.StatusBadRequest, err.Error()+"\n")
		return
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.String(http.StatusBadRequest, err.Error()+"\n")
		return
	}

	member, err := c.store.Update(request.Id, &request.MemberUpdateDTO)
	if err != nil {
		log.Printf("error updating member: %v", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, member.ToResponseDTO())
}
