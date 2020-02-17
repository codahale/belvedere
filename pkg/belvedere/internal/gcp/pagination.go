package gcp

// Paginate pages over a function, passing in the current page token and receiving the next page
// token.
func Paginate(f func(string) (string, error)) error {
	pageToken := ""
	for {
		s, err := f(pageToken)
		if err != nil {
			return err
		}
		if s == "" {
			return nil
		}
		pageToken = s
	}
}
