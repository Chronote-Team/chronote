package app

func canAccess(userID, authorID uint, visibility string) bool {
	if userID == authorID {
		return true
	}
	return visibility == "public"
}
