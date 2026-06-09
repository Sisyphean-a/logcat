package app

func (c *Controller) markDirtyLocked() {
	c.revision++
}
