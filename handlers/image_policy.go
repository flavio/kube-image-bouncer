package handlers

import (
	"net/http"

	"github.com/flavio/kube-image-bouncer/rules"

	"github.com/labstack/echo"
	"k8s.io/api/imagepolicy/v1alpha1"
)

func PostImagePolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		var imageReview v1alpha1.ImageReview
		var review v1alpha1.ImageReview

		// Map imcoming JSON body to the new Entry
		err := c.Bind(&imageReview)
		if err != nil {
			c.Logger().Errorf("Something went wrong while unmarshalling: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		c.Logger().Debugf("image review request: %+v", imageReview)

		allow := true
		images := []string{}

		for _, container := range imageReview.Spec.Containers {
			images = append(images, container.Image)
			usingLatest, err := rules.IsUsingLatestTag(container.Image)
			if err != nil {
				c.Logger().Errorf("Error while parsing image name: %+v", err)
				return c.JSON(http.StatusInternalServerError, "error while parsing image name")
			}
			if usingLatest {
				allow = false
				review.Status.Reason = "Images using latest tag are not allowed"
				break
			}
		}

		review.Status.Allowed = allow

		if allow {
			c.Logger().Debugf("All images accepted: %v", images)
		} else {
			c.Logger().Infof("Rejected images: %v", images)
		}

		c.Logger().Debugf("reply: %+v", review)

		return c.JSON(http.StatusOK, review)
	}
}
