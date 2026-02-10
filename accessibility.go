package fyne

type AccessibleRole string

const (
	AccessibleRoleButton    AccessibleRole = "button"
	AccessibleRoleContainer AccessibleRole = "container"
	AccessibleRoleLink      AccessibleRole = "link"
	AccessibleRoleText      AccessibleRole = "text"
)

type Accessible interface {
	AccessibilityLabel() string
	AccessibilityRole() AccessibleRole
}
