// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/session"
)

func (h *handler) clearFeedEntries(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	if !h.store.FeedExists(userID, feedID) {
		html.NotFound(w, r)
		return
	}

	if err := h.store.RemoveFeedEntries(userID, feedID); err != nil {
		html.ServerError(w, r, err)
		return
	}

	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))
	sess.NewFlashMessage(printer.Print("alert.feed_entries_cleared"))

	html.Redirect(w, r, route.Path(h.router, "editFeed", "feedID", feedID))
}
