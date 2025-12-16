package group

type GroupMember struct {
	UserID   string `bson:"user_id,omitempty" json:"user_id"`
	IsAdmin  bool   `bson:"is_admin,omitempty" json:"is_admin"`
	JoinedAt string `bson:"joined_at,omitempty" json:"joined_at"` // Stored as string
}

type Group struct {
	ID          string        `bson:"_id,omitempty" json:"_id"`
	Title       string        `bson:"title,omitempty" json:"title"`
	Description string        `bson:"description,omitempty" json:"description"`
	GroupIcon   string        `bson:"group_icon,omitempty" json:"group_icon"`   // URL or file path of group icon
	CreatedBy   string        `bson:"created_by,omitempty" json:"created_by"`   // User ID of the creator
	CreatedAt   string        `bson:"created_at,omitempty" json:"created_at"`   // Stored as string
	UpdatedAt   string        `bson:"updated_at,omitempty" json:"updated_at"`   // Stored as string
	DisappearingMsg int       `bson:"disappearing_msg,omitempty" json:"disappearing_msg"`
	MemberCanEdit bool 		  `bson:"member_can_edit,omitempty" json:"member_can_edit"`
	MemberCanSend bool 		  `bson:"member_can_send,omitempty" json:"member_can_send"`
	MemberCanAdd  bool 		  `bson:"member_can_add,omitempty" json:"member_can_add"`
	AdminApprove  bool 		  `bson:"admin_approve,omitempty" json:"admin_approve"`
	Members     []GroupMember `bson:"members,omitempty" json:"members"`
}
