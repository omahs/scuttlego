package badger_test

import (
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestSocialGraphRepository_RemoveDropsContactDataForTheSpecifiedFeed(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	iden1 := fixtures.SomeRefIdentity()
	iden2 := fixtures.SomeRefIdentity()

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden1, fixtures.SomeRefIdentity(), known.ContactActionFollow)
		applyContactAction(t, adapters, iden2, fixtures.SomeRefIdentity(), known.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		contacts, err = adapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err = adapters.SocialGraphRepository.Remove(iden1)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden1)
		require.NoError(t, err)
		require.Empty(t, contacts)

		contacts, err = adapters.SocialGraphRepository.GetContacts(iden2)
		require.NoError(t, err)
		require.NotEmpty(t, contacts)

		return nil
	})
	require.NoError(t, err)
}

func TestSocialGraphRepository_GetContacts(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	iden := fixtures.SomeRefIdentity()

	target1 := fixtures.SomeRefIdentity()
	target2 := fixtures.SomeRefIdentity()
	target3 := fixtures.SomeRefIdentity()

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden, target1, known.ContactActionFollow)
		applyContactAction(t, adapters, iden, target2, known.ContactActionFollow)
		applyContactAction(t, adapters, iden, target3, known.ContactActionFollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(iden, target1, true, false),
				feeds.MustNewContactFromHistory(iden, target2, true, false),
				feeds.MustNewContactFromHistory(iden, target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		applyContactAction(t, adapters, iden, target1, known.ContactActionBlock)
		applyContactAction(t, adapters, iden, target2, known.ContactActionUnfollow)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		contacts, err := adapters.SocialGraphRepository.GetContacts(iden)
		require.NoError(t, err)
		sortAndRequireEqualContacts(
			t,
			[]*feeds.Contact{
				feeds.MustNewContactFromHistory(iden, target1, true, true),
				feeds.MustNewContactFromHistory(iden, target2, false, false),
				feeds.MustNewContactFromHistory(iden, target3, true, false),
			},
			contacts,
		)

		return nil
	})
	require.NoError(t, err)
}

func sortAndRequireEqualContacts(t *testing.T, a []*feeds.Contact, b []*feeds.Contact) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Target().String() < a[j].Target().String()
	})

	sort.Slice(b, func(i, j int) bool {
		return b[i].Target().String() < b[j].Target().String()
	})

	require.Equal(t, a, b)
}

func applyContactAction(t *testing.T, adapters badger.TestAdapters, a refs.Identity, b refs.Identity, action known.ContactAction) {
	err := adapters.SocialGraphRepository.UpdateContact(a, b, func(contact *feeds.Contact) error {
		return contact.Update(known.MustNewContactActions([]known.ContactAction{action}))
	})
	require.NoError(t, err)
}
