package repository

import (
	"context"
	"sort"
	"testing"

	"github.com/project/library/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryImpl_CreateAuthor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		author      entity.Author
		preSetup    func(repo *inMemoryImpl)
		want        entity.Author
		wantErr     bool
		expectedErr error
	}{
		{
			name: "create author | successful creation",
			author: entity.Author{
				ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
				Name: "Test Author",
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			want: entity.Author{
				ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
				Name: "Test Author",
			},
			wantErr: false,
		},
		{
			name: "create author | author already exists",
			author: entity.Author{
				ID:   "existing-author-id",
				Name: "New Author Name",
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["existing-author-id"] = &entity.Author{
					ID:   "existing-author-id",
					Name: "Existing Author",
				}
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorAlreadyExists,
		},
		{
			name: "create author | multiple authors",
			author: entity.Author{
				ID:   "new-author-id",
				Name: "New Author",
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["author-1"] = &entity.Author{
					ID:   "author-1",
					Name: "Author 1",
				}
				repo.authors["author-2"] = &entity.Author{
					ID:   "author-2",
					Name: "Author 2",
				}
			},
			want: entity.Author{
				ID:   "new-author-id",
				Name: "New Author",
			},
			wantErr: false,
		},
		{
			name: "create author | empty name",
			author: entity.Author{
				ID:   "empty-name-author",
				Name: "",
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			want: entity.Author{
				ID:   "empty-name-author",
				Name: "",
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			got, err := repo.CreateAuthor(ctx, test.author)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want.ID, got.ID)
				assert.Equal(t, test.want.Name, got.Name)

				// Проверка, что автор в репе.
				repo.authorsMx.RLock()
				defer repo.authorsMx.RUnlock()
				storedAuthor, exists := repo.authors[test.author.ID]
				assert.True(t, exists)
				assert.Equal(t, test.want.ID, storedAuthor.ID)
				assert.Equal(t, test.want.Name, storedAuthor.Name)
			}
		})
	}
}

func TestInMemoryImpl_UpdateAuthor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		authorID    string
		authorName  string
		preSetup    func(repo *inMemoryImpl)
		wantErr     bool
		expectedErr error
		verify      func(t *testing.T, repo *inMemoryImpl)
	}{
		{
			name:       "update author | successful update",
			authorID:   "7a948d89-108c-4133-be30-788bd453c0cd",
			authorName: "Updated Author Name",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["7a948d89-108c-4133-be30-788bd453c0cd"] = &entity.Author{
					ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
					Name: "Original Author Name",
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.authorsMx.RLock()
				defer repo.authorsMx.RUnlock()
				author, exists := repo.authors["7a948d89-108c-4133-be30-788bd453c0cd"]
				assert.True(t, exists)
				assert.Equal(t, "Updated Author Name", author.Name)
				assert.Equal(t, "7a948d89-108c-4133-be30-788bd453c0cd", author.ID)
			},
		},
		{
			name:       "update author | author not found",
			authorID:   "non-existent-id",
			authorName: "Updated Name",
			preSetup: func(repo *inMemoryImpl) {
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.authorsMx.RLock()
				defer repo.authorsMx.RUnlock()
				_, exists := repo.authors["non-existent-id"]
				assert.False(t, exists)
			},
		},
		{
			name:       "update author | update with same name",
			authorID:   "author-same-name",
			authorName: "Same Name",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["author-same-name"] = &entity.Author{
					ID:   "author-same-name",
					Name: "Same Name",
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.authorsMx.RLock()
				defer repo.authorsMx.RUnlock()
				author, exists := repo.authors["author-same-name"]
				assert.True(t, exists)
				assert.Equal(t, "Same Name", author.Name)
			},
		},
		{
			name:       "update author | update with multiple authors present",
			authorID:   "author-to-update",
			authorName: "Updated Name",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["author-to-update"] = &entity.Author{
					ID:   "author-to-update",
					Name: "Original Name",
				}
				repo.authors["other-author-1"] = &entity.Author{
					ID:   "other-author-1",
					Name: "Other Author 1",
				}
				repo.authors["other-author-2"] = &entity.Author{
					ID:   "other-author-2",
					Name: "Other Author 2",
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.authorsMx.RLock()
				defer repo.authorsMx.RUnlock()

				author, exists := repo.authors["author-to-update"]
				assert.True(t, exists)
				assert.Equal(t, "Updated Name", author.Name)

				other1, exists := repo.authors["other-author-1"]
				assert.True(t, exists)
				assert.Equal(t, "Other Author 1", other1.Name)

				other2, exists := repo.authors["other-author-2"]
				assert.True(t, exists)
				assert.Equal(t, "Other Author 2", other2.Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			err := repo.UpdateAuthor(ctx, test.authorID, test.authorName)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			if test.verify != nil {
				test.verify(t, repo)
			}
		})
	}
}

func TestInMemoryImpl_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		authorID    string
		preSetup    func(repo *inMemoryImpl)
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "get author info | successful get",
			authorID: "7a948d89-108c-4133-be30-788bd453c0cd",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["7a948d89-108c-4133-be30-788bd453c0cd"] = &entity.Author{
					ID:   "7a948d89-108c-4133-be30-788bd453c0cd",
					Name: "Test Author",
				}
			},
			want:    "Test Author",
			wantErr: false,
		},
		{
			name:     "get author info | author not found",
			authorID: "non-existent-id",
			preSetup: func(repo *inMemoryImpl) {
			},
			wantErr:     true,
			expectedErr: entity.ErrAuthorNotFound,
		},
		{
			name:     "get author info | with multiple authors",
			authorID: "target-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["target-author"] = &entity.Author{
					ID:   "target-author",
					Name: "Target Author",
				}
				repo.authors["other-author-1"] = &entity.Author{
					ID:   "other-author-1",
					Name: "Other Author 1",
				}
				repo.authors["other-author-2"] = &entity.Author{
					ID:   "other-author-2",
					Name: "Other Author 2",
				}
			},
			want:    "Target Author",
			wantErr: false,
		},
		{
			name:     "get author info | author with special characters",
			authorID: "special-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["special-author"] = &entity.Author{
					ID:   "special-author",
					Name: "Author with 🚀 Emoji",
				}
			},
			want:    "Author with 🚀 Emoji",
			wantErr: false,
		},
		{
			name:     "get author info | author with long name",
			authorID: "long-name-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.authors["long-name-author"] = &entity.Author{
					ID:   "long-name-author",
					Name: "Very Long Author Name With Many Words And Special Characters!@#$%^&*()",
				}
			},
			want:    "Very Long Author Name With Many Words And Special Characters!@#$%^&*()",
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			got, err := repo.GetAuthorInfo(ctx, test.authorID)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestInMemoryImpl_CreateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		book        entity.Book
		preSetup    func(repo *inMemoryImpl)
		want        entity.Book
		wantErr     bool
		expectedErr error
		verify      func(t *testing.T, repo *inMemoryImpl)
	}{
		{
			name: "create book | successful creation without authors",
			book: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{},
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			want: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{},
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["7a948d89-108c-4133-be30-788bd453c0cd"]
				assert.True(t, exists)
				assert.Equal(t, "Test Book", storedBook.Name)
				assert.Equal(t, []string{}, storedBook.AuthorIDs)
			},
		},
		{
			name: "create book | successful creation with authors",
			book: entity.Book{
				ID:        "book-with-authors",
				Name:      "Book With Authors",
				AuthorIDs: []string{"author-1", "author-2"},
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			want: entity.Book{
				ID:        "book-with-authors",
				Name:      "Book With Authors",
				AuthorIDs: []string{"author-1", "author-2"},
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["book-with-authors"]
				assert.True(t, exists)
				assert.Equal(t, "Book With Authors", storedBook.Name)
				assert.Equal(t, []string{"author-1", "author-2"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "create book | book already exists",
			book: entity.Book{
				ID:        "existing-book-id",
				Name:      "New Book Name",
				AuthorIDs: []string{"author-1"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["existing-book-id"] = &entity.Book{
					ID:        "existing-book-id",
					Name:      "Existing Book",
					AuthorIDs: []string{"existing-author"},
				}
			},
			wantErr:     true,
			expectedErr: entity.ErrBookAlreadyExists,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["existing-book-id"]
				assert.True(t, exists)
				assert.Equal(t, "Existing Book", storedBook.Name)
				assert.Equal(t, []string{"existing-author"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "create book | multiple books",
			book: entity.Book{
				ID:        "new-book-id",
				Name:      "New Book",
				AuthorIDs: []string{"author-1"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["book-1"] = &entity.Book{
					ID:        "book-1",
					Name:      "Book 1",
					AuthorIDs: []string{"author-1"},
				}
				repo.books["book-2"] = &entity.Book{
					ID:        "book-2",
					Name:      "Book 2",
					AuthorIDs: []string{"author-2"},
				}
			},
			want: entity.Book{
				ID:        "new-book-id",
				Name:      "New Book",
				AuthorIDs: []string{"author-1"},
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()

				newBook, exists := repo.books["new-book-id"]
				assert.True(t, exists)
				assert.Equal(t, "New Book", newBook.Name)

				book1, exists := repo.books["book-1"]
				assert.True(t, exists)
				assert.Equal(t, "Book 1", book1.Name)
				book2, exists := repo.books["book-2"]
				assert.True(t, exists)
				assert.Equal(t, "Book 2", book2.Name)
			},
		},
		{
			name: "create book | empty name",
			book: entity.Book{
				ID:        "empty-name-book",
				Name:      "",
				AuthorIDs: []string{"author-1"},
			},
			preSetup: func(repo *inMemoryImpl) {

			},
			want: entity.Book{
				ID:        "empty-name-book",
				Name:      "",
				AuthorIDs: []string{"author-1"},
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["empty-name-book"]
				assert.True(t, exists)
				assert.Equal(t, "", storedBook.Name)
				assert.Equal(t, []string{"author-1"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "create book | empty authors list",
			book: entity.Book{
				ID:        "no-authors-book",
				Name:      "No Authors Book",
				AuthorIDs: []string{},
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			want: entity.Book{
				ID:        "no-authors-book",
				Name:      "No Authors Book",
				AuthorIDs: []string{},
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["no-authors-book"]
				assert.True(t, exists)
				assert.Equal(t, "No Authors Book", storedBook.Name)
				assert.Equal(t, []string{}, storedBook.AuthorIDs)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			got, err := repo.CreateBook(ctx, test.book)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want.ID, got.ID)
				assert.Equal(t, test.want.Name, got.Name)
				assert.Equal(t, test.want.AuthorIDs, got.AuthorIDs)
			}

			if test.verify != nil {
				test.verify(t, repo)
			}
		})
	}
}

func TestInMemoryImpl_GetBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		bookID      string
		preSetup    func(repo *inMemoryImpl)
		want        entity.Book
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "get book | successful get without authors",
			bookID: "7a948d89-108c-4133-be30-788bd453c0cd",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["7a948d89-108c-4133-be30-788bd453c0cd"] = &entity.Book{
					ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
					Name:      "Test Book",
					AuthorIDs: []string{},
				}
			},
			want: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Test Book",
				AuthorIDs: []string{},
			},
			wantErr: false,
		},
		{
			name:   "get book | successful get with authors",
			bookID: "book-with-authors",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["book-with-authors"] = &entity.Book{
					ID:        "book-with-authors",
					Name:      "Book With Authors",
					AuthorIDs: []string{"author-1", "author-2"},
				}
			},
			want: entity.Book{
				ID:        "book-with-authors",
				Name:      "Book With Authors",
				AuthorIDs: []string{"author-1", "author-2"},
			},
			wantErr: false,
		},
		{
			name:   "get book | book not found",
			bookID: "non-existent-id",
			preSetup: func(repo *inMemoryImpl) {
			},
			wantErr:     true,
			expectedErr: entity.ErrBookNotFound,
		},
		{
			name:   "get book | book with empty name",
			bookID: "empty-name-book-id",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["empty-name-book-id"] = &entity.Book{
					ID:        "empty-name-book-id",
					Name:      "",
					AuthorIDs: []string{"author-1"},
				}
			},
			want: entity.Book{
				ID:        "empty-name-book-id",
				Name:      "",
				AuthorIDs: []string{"author-1"},
			},
			wantErr: false,
		},
		{
			name:   "get book | book with empty authors",
			bookID: "no-authors-book-id",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["no-authors-book-id"] = &entity.Book{
					ID:        "no-authors-book-id",
					Name:      "No Authors Book",
					AuthorIDs: []string{},
				}
			},
			want: entity.Book{
				ID:        "no-authors-book-id",
				Name:      "No Authors Book",
				AuthorIDs: []string{},
			},
			wantErr: false,
		},
		{
			name:   "get book | with multiple books present",
			bookID: "target-book",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["target-book"] = &entity.Book{
					ID:        "target-book",
					Name:      "Target Book",
					AuthorIDs: []string{"target-author"},
				}
				repo.books["other-book-1"] = &entity.Book{
					ID:        "other-book-1",
					Name:      "Other Book 1",
					AuthorIDs: []string{"other-author-1"},
				}
				repo.books["other-book-2"] = &entity.Book{
					ID:        "other-book-2",
					Name:      "Other Book 2",
					AuthorIDs: []string{"other-author-2"},
				}
			},
			want: entity.Book{
				ID:        "target-book",
				Name:      "Target Book",
				AuthorIDs: []string{"target-author"},
			},
			wantErr: false,
		},
		{
			name:   "get book | book with special characters",
			bookID: "special-book",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["special-book"] = &entity.Book{
					ID:        "special-book",
					Name:      "Book with 🚀 Emoji and #Hashtag!",
					AuthorIDs: []string{"special-author"},
				}
			},
			want: entity.Book{
				ID:        "special-book",
				Name:      "Book with 🚀 Emoji and #Hashtag!",
				AuthorIDs: []string{"special-author"},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			got, err := repo.GetBook(ctx, test.bookID)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want.ID, got.ID)
				assert.Equal(t, test.want.Name, got.Name)
				assert.Equal(t, test.want.AuthorIDs, got.AuthorIDs)
			}
		})
	}
}

func TestInMemoryImpl_UpdateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		book        entity.Book
		preSetup    func(repo *inMemoryImpl)
		wantErr     bool
		expectedErr error
		verify      func(t *testing.T, repo *inMemoryImpl)
	}{
		{
			name: "update book | successful update name and authors",
			book: entity.Book{
				ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
				Name:      "Updated Book Name",
				AuthorIDs: []string{"new-author-1", "new-author-2"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["7a948d89-108c-4133-be30-788bd453c0cd"] = &entity.Book{
					ID:        "7a948d89-108c-4133-be30-788bd453c0cd",
					Name:      "Original Book Name",
					AuthorIDs: []string{"old-author"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["7a948d89-108c-4133-be30-788bd453c0cd"]
				assert.True(t, exists)
				assert.Equal(t, "Updated Book Name", storedBook.Name)
				assert.Equal(t, []string{"new-author-1", "new-author-2"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "update book | successful update name only",
			book: entity.Book{
				ID:        "update-name-only",
				Name:      "New Name",
				AuthorIDs: []string{"same-author"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["update-name-only"] = &entity.Book{
					ID:        "update-name-only",
					Name:      "Old Name",
					AuthorIDs: []string{"same-author"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["update-name-only"]
				assert.True(t, exists)
				assert.Equal(t, "New Name", storedBook.Name)
				assert.Equal(t, []string{"same-author"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "update book | successful update authors only",
			book: entity.Book{
				ID:        "update-authors-only",
				Name:      "Same Name",
				AuthorIDs: []string{"new-author-1"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["update-authors-only"] = &entity.Book{
					ID:        "update-authors-only",
					Name:      "Same Name",
					AuthorIDs: []string{"old-author"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["update-authors-only"]
				assert.True(t, exists)
				assert.Equal(t, "Same Name", storedBook.Name)
				assert.Equal(t, []string{"new-author-1"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "update book | book not found",
			book: entity.Book{
				ID:        "non-existent-id",
				Name:      "Some Name",
				AuthorIDs: []string{"some-author"},
			},
			preSetup: func(repo *inMemoryImpl) {
			},
			wantErr:     true,
			expectedErr: entity.ErrBookNotFound,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				_, exists := repo.books["non-existent-id"]
				assert.False(t, exists)
			},
		},
		{
			name: "update book | update to empty name",
			book: entity.Book{
				ID:        "empty-name-update",
				Name:      "",
				AuthorIDs: []string{"author-1"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["empty-name-update"] = &entity.Book{
					ID:        "empty-name-update",
					Name:      "Original Name",
					AuthorIDs: []string{"author-1"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["empty-name-update"]
				assert.True(t, exists)
				assert.Equal(t, "", storedBook.Name)
				assert.Equal(t, []string{"author-1"}, storedBook.AuthorIDs)
			},
		},
		{
			name: "update book | update to empty authors",
			book: entity.Book{
				ID:        "empty-authors-update",
				Name:      "Book Name",
				AuthorIDs: []string{},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["empty-authors-update"] = &entity.Book{
					ID:        "empty-authors-update",
					Name:      "Book Name",
					AuthorIDs: []string{"old-author"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()
				storedBook, exists := repo.books["empty-authors-update"]
				assert.True(t, exists)
				assert.Equal(t, "Book Name", storedBook.Name)
				assert.Equal(t, []string{}, storedBook.AuthorIDs)
			},
		},
		{
			name: "update book | with multiple books present",
			book: entity.Book{
				ID:        "target-book",
				Name:      "Updated Target Book",
				AuthorIDs: []string{"updated-author"},
			},
			preSetup: func(repo *inMemoryImpl) {
				repo.books["target-book"] = &entity.Book{
					ID:        "target-book",
					Name:      "Original Target Book",
					AuthorIDs: []string{"original-author"},
				}
				repo.books["other-book-1"] = &entity.Book{
					ID:        "other-book-1",
					Name:      "Other Book 1",
					AuthorIDs: []string{"other-author-1"},
				}
				repo.books["other-book-2"] = &entity.Book{
					ID:        "other-book-2",
					Name:      "Other Book 2",
					AuthorIDs: []string{"other-author-2"},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, repo *inMemoryImpl) {
				repo.booksMx.RLock()
				defer repo.booksMx.RUnlock()

				targetBook, exists := repo.books["target-book"]
				assert.True(t, exists)
				assert.Equal(t, "Updated Target Book", targetBook.Name)
				assert.Equal(t, []string{"updated-author"}, targetBook.AuthorIDs)

				otherBook1, exists := repo.books["other-book-1"]
				assert.True(t, exists)
				assert.Equal(t, "Other Book 1", otherBook1.Name)
				otherBook2, exists := repo.books["other-book-2"]
				assert.True(t, exists)
				assert.Equal(t, "Other Book 2", otherBook2.Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			err := repo.UpdateBook(ctx, test.book)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			if test.verify != nil {
				test.verify(t, repo)
			}
		})
	}
}

func TestInMemoryImpl_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		authorID    string
		preSetup    func(repo *inMemoryImpl)
		want        []entity.Book
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "get author books | author has multiple books",
			authorID: "target-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["book-1"] = &entity.Book{
					ID:        "book-1",
					Name:      "Book 1",
					AuthorIDs: []string{"target-author"},
				}
				repo.books["book-2"] = &entity.Book{
					ID:        "book-2",
					Name:      "Book 2",
					AuthorIDs: []string{"target-author", "other-author"},
				}
				repo.books["book-3"] = &entity.Book{
					ID:        "book-3",
					Name:      "Book 3",
					AuthorIDs: []string{"target-author"},
				}
			},
			want: []entity.Book{
				{
					ID:        "book-1",
					Name:      "Book 1",
					AuthorIDs: []string{"target-author"},
				},
				{
					ID:        "book-2",
					Name:      "Book 2",
					AuthorIDs: []string{"target-author", "other-author"},
				},
				{
					ID:        "book-3",
					Name:      "Book 3",
					AuthorIDs: []string{"target-author"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books | author has single book",
			authorID: "single-book-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["single-book"] = &entity.Book{
					ID:        "single-book",
					Name:      "Single Book",
					AuthorIDs: []string{"single-book-author"},
				}
				repo.books["other-book"] = &entity.Book{
					ID:        "other-book",
					Name:      "Other Book",
					AuthorIDs: []string{"different-author"},
				}
			},
			want: []entity.Book{
				{
					ID:        "single-book",
					Name:      "Single Book",
					AuthorIDs: []string{"single-book-author"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books | author has no books",
			authorID: "author-with-no-books",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["book-1"] = &entity.Book{
					ID:        "book-1",
					Name:      "Book 1",
					AuthorIDs: []string{"other-author-1"},
				}
				repo.books["book-2"] = &entity.Book{
					ID:        "book-2",
					Name:      "Book 2",
					AuthorIDs: []string{"other-author-2"},
				}
			},
			want:    []entity.Book{},
			wantErr: false,
		},
		{
			name:     "get author books | empty repository",
			authorID: "any-author",
			preSetup: func(repo *inMemoryImpl) {
				// Empty repository
			},
			want:    []entity.Book{},
			wantErr: false,
		},
		{
			name:     "get author books | author in multiple books with different co-authors",
			authorID: "main-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["solo-book"] = &entity.Book{
					ID:        "solo-book",
					Name:      "Solo Book",
					AuthorIDs: []string{"main-author"},
				}
				repo.books["duo-book"] = &entity.Book{
					ID:        "duo-book",
					Name:      "Duo Book",
					AuthorIDs: []string{"main-author", "co-author-1"},
				}
				repo.books["trio-book"] = &entity.Book{
					ID:        "trio-book",
					Name:      "Trio Book",
					AuthorIDs: []string{"main-author", "co-author-1", "co-author-2"},
				}
			},
			want: []entity.Book{
				{
					ID:        "solo-book",
					Name:      "Solo Book",
					AuthorIDs: []string{"main-author"},
				},
				{
					ID:        "duo-book",
					Name:      "Duo Book",
					AuthorIDs: []string{"main-author", "co-author-1"},
				},
				{
					ID:        "trio-book",
					Name:      "Trio Book",
					AuthorIDs: []string{"main-author", "co-author-1", "co-author-2"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books | books with empty authors",
			authorID: "some-author",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["no-authors-book"] = &entity.Book{
					ID:        "no-authors-book",
					Name:      "No Authors Book",
					AuthorIDs: []string{},
				}
				repo.books["with-authors-book"] = &entity.Book{
					ID:        "with-authors-book",
					Name:      "With Authors Book",
					AuthorIDs: []string{"some-author"},
				}
			},
			want: []entity.Book{
				{
					ID:        "with-authors-book",
					Name:      "With Authors Book",
					AuthorIDs: []string{"some-author"},
				},
			},
			wantErr: false,
		},
		{
			name:     "get author books | case sensitivity in author IDs",
			authorID: "exact-author-id",
			preSetup: func(repo *inMemoryImpl) {
				repo.books["exact-match"] = &entity.Book{
					ID:        "exact-match",
					Name:      "Exact Match Book",
					AuthorIDs: []string{"exact-author-id"},
				}
				repo.books["different-case"] = &entity.Book{
					ID:        "different-case",
					Name:      "Different Case Book",
					AuthorIDs: []string{"Exact-Author-ID"}, // Different case
				}
				repo.books["similar"] = &entity.Book{
					ID:        "similar",
					Name:      "Similar Book",
					AuthorIDs: []string{"exact-author-id-extra"}, // Similar but different
				}
			},
			want: []entity.Book{
				{
					ID:        "exact-match",
					Name:      "Exact Match Book",
					AuthorIDs: []string{"exact-author-id"},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.preSetup != nil {
				test.preSetup(repo)
			}

			got, err := repo.GetAuthorBooks(ctx, test.authorID)

			if test.wantErr {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(test.want), len(got))

				sort.Slice(test.want, func(i, j int) bool {
					return test.want[i].ID < test.want[j].ID
				})
				sort.Slice(got, func(i, j int) bool {
					return got[i].ID < got[j].ID
				})

				for i, expectedBook := range test.want {
					assert.Equal(t, expectedBook.ID, got[i].ID)
					assert.Equal(t, expectedBook.Name, got[i].Name)
					assert.Equal(t, expectedBook.AuthorIDs, got[i].AuthorIDs)
				}
			}
		})
	}
}
