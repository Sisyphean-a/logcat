package app

func (c *Controller) markDirtyLocked() {
	c.revision++
	select {
	case c.dirtyCh <- struct{}{}:
	default:
	}
}
