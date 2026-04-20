package user

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	userdomain "github.com/jrmarcello/gopherplate/internal/domain/user"
	"github.com/jrmarcello/gopherplate/internal/domain/user/vo"
	"github.com/jrmarcello/gopherplate/internal/usecases/user/dto"
	"github.com/jrmarcello/gopherplate/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TC-UC-40: GetUseCase cache hit emits cache.hit event with cache.key.
func TestGetUseCase_Events_CacheHit(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := "018e4a2c-6b4d-7000-9410-abcdef123456"
	cacheKey := "user:" + id

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*dto.GetOutput")).
		Return(nil)

	uc := NewGetUseCase(mockRepo).WithCache(mockCache)
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.GetInput{ID: id})
	assert.NoError(t, executeErr)

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.hit"), "expected cache.hit event; got %v", eventNames(stub))
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.hit", "cache.key"))
	assert.False(t, hasEvent(stub, "cache.miss"), "cache.miss must not be emitted on hit")
	mockRepo.AssertNotCalled(t, "FindByID")
}

// TC-UC-41: GetUseCase cache miss -> DB hit -> cache.set emits both events.
func TestGetUseCase_Events_CacheMissThenSet(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*dto.GetOutput")).
		Return(errors.New("cache miss"))
	mockRepo.On("FindByID", mock.Anything, id).Return(entity, nil)
	mockCache.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("*dto.GetOutput")).
		Return(nil)

	uc := NewGetUseCase(mockRepo).WithCache(mockCache)
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.GetInput{ID: id.String()})
	assert.NoError(t, executeErr)

	stub := finalize()
	names := eventNames(stub)
	assert.Contains(t, names, "cache.miss")
	assert.Contains(t, names, "cache.set")
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.miss", "cache.key"))
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.set", "cache.key"))
	// Order: miss should precede set
	missIdx, setIdx := -1, -1
	for i, n := range names {
		if n == "cache.miss" {
			missIdx = i
		}
		if n == "cache.set" {
			setIdx = i
		}
	}
	assert.True(t, missIdx >= 0 && setIdx > missIdx, "cache.miss must precede cache.set; got %v", names)
}

// TC-UC-42: cache Set failure emits cache.set_failed; span status stays Ok.
func TestGetUseCase_Events_CacheSetFailed(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*dto.GetOutput")).
		Return(errors.New("cache miss"))
	mockRepo.On("FindByID", mock.Anything, id).Return(entity, nil)
	mockCache.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("*dto.GetOutput")).
		Return(errors.New("redis down"))

	uc := NewGetUseCase(mockRepo).WithCache(mockCache)
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.GetInput{ID: id.String()})
	assert.NoError(t, executeErr, "cache set failure must not surface as an error")

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.set_failed"),
		"expected cache.set_failed event; got %v", eventNames(stub))
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.set_failed", "cache.key"))
	assert.Equal(t, "redis down", eventAttr(stub, "cache.set_failed", "error.message"))
	assert.False(t, hasEvent(stub, "cache.set"), "cache.set must not fire when Set fails")
	// span status must remain Ok (cache set failure is a warning, not a failure)
	assert.NotEqual(t, "Error", stub.Status.Code.String(),
		"span must not be marked Error on cache set failure")
}

// TC-UC-43: singleflight shared path — second caller joining the first emits
// singleflight.shared on its own span.
func TestGetUseCase_Events_SingleflightShared(t *testing.T) {
	mockRepo := new(MockRepository)

	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	// block the first call long enough for the second to join
	release := make(chan struct{})
	started := make(chan struct{}, 1)
	mockRepo.On("FindByID", mock.Anything, id).
		Run(func(mock.Arguments) {
			select {
			case started <- struct{}{}:
			default:
			}
			<-release
		}).
		Return(entity, nil)

	fg := cache.NewFlightGroup()
	uc := NewGetUseCase(mockRepo).WithFlight(fg)

	ctx1, finalize1 := newRecordingSpanContext(t)
	ctx2, finalize2 := newRecordingSpanContext(t)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, callErr := uc.Execute(ctx1, dto.GetInput{ID: id.String()})
		assert.NoError(t, callErr)
	}()

	// wait for first call to enter singleflight
	<-started

	go func() {
		defer wg.Done()
		// small delay to ensure the second caller joins while the first is in-flight
		time.Sleep(10 * time.Millisecond)
		_, callErr := uc.Execute(ctx2, dto.GetInput{ID: id.String()})
		assert.NoError(t, callErr)
	}()

	time.Sleep(30 * time.Millisecond)
	close(release)
	wg.Wait()

	stub1 := finalize1()
	stub2 := finalize2()

	// Go's x/sync/singleflight returns shared=true to EVERY caller that
	// shared a single execution, including the leader. Both callers benefited
	// from the dedup, so both spans carry the event. The load-bearing
	// invariant is that the repo was called exactly once despite two
	// concurrent callers — proving the dedup actually happened.
	assert.True(t, hasEvent(stub1, "singleflight.shared"),
		"expected singleflight.shared on stub1; got %v", eventNames(stub1))
	assert.True(t, hasEvent(stub2, "singleflight.shared"),
		"expected singleflight.shared on stub2; got %v", eventNames(stub2))
	mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
}

// TC-UC-44: UpdateUseCase cache invalidation success emits cache.invalidated.
func TestUpdateUseCase_Events_CacheInvalidated(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockRepo.On("FindByID", mock.Anything, id).Return(entity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(nil)

	uc := NewUpdateUseCase(mockRepo).WithCache(mockCache)
	newName := "João Silva Updated"
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.UpdateInput{ID: id.String(), Name: &newName})
	assert.NoError(t, executeErr)

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.invalidated"),
		"expected cache.invalidated event; got %v", eventNames(stub))
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.invalidated", "cache.key"))
}

// TC-UC-45: UpdateUseCase cache Delete failure emits cache.invalidate_failed;
// span status stays Ok.
func TestUpdateUseCase_Events_CacheInvalidateFailed(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockRepo.On("FindByID", mock.Anything, id).Return(entity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(errors.New("redis unreachable"))

	uc := NewUpdateUseCase(mockRepo).WithCache(mockCache)
	newName := "João Silva Updated"
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.UpdateInput{ID: id.String(), Name: &newName})
	assert.NoError(t, executeErr)

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.invalidate_failed"),
		"expected cache.invalidate_failed event; got %v", eventNames(stub))
	assert.Equal(t, "redis unreachable", eventAttr(stub, "cache.invalidate_failed", "error.message"))
	assert.NotEqual(t, "Error", stub.Status.Code.String(),
		"span must not be marked Error on cache invalidation failure")
}

// TC-UC-46: DeleteUseCase cache invalidation success emits cache.invalidated.
func TestDeleteUseCase_Events_CacheInvalidated(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()

	mockRepo.On("Delete", mock.Anything, id).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(nil)

	uc := NewDeleteUseCase(mockRepo).WithCache(mockCache)
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.DeleteInput{ID: id.String()})
	assert.NoError(t, executeErr)

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.invalidated"),
		"expected cache.invalidated event; got %v", eventNames(stub))
	assert.Equal(t, cacheKey, eventAttr(stub, "cache.invalidated", "cache.key"))
}

// TC-UC-47: DeleteUseCase cache Delete failure emits cache.invalidate_failed.
func TestDeleteUseCase_Events_CacheInvalidateFailed(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	id := vo.NewID()
	cacheKey := "user:" + id.String()

	mockRepo.On("Delete", mock.Anything, id).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(errors.New("redis unreachable"))

	uc := NewDeleteUseCase(mockRepo).WithCache(mockCache)
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.DeleteInput{ID: id.String()})
	assert.NoError(t, executeErr)

	stub := finalize()
	assert.True(t, hasEvent(stub, "cache.invalidate_failed"),
		"expected cache.invalidate_failed event; got %v", eventNames(stub))
	assert.Equal(t, "redis unreachable", eventAttr(stub, "cache.invalidate_failed", "error.message"))
}

// TC-UC-48: GetUseCase with no cache wired — no cache.* events emitted.
// Ensures we don't regress into spurious event emission when cache is absent.
func TestGetUseCase_Events_NoCache(t *testing.T) {
	mockRepo := new(MockRepository)

	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")
	entity := &userdomain.User{ID: id, Name: "João", Email: email, Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockRepo.On("FindByID", mock.Anything, id).Return(entity, nil)

	uc := NewGetUseCase(mockRepo) // no WithCache
	ctx, finalize := newRecordingSpanContext(t)

	_, executeErr := uc.Execute(ctx, dto.GetInput{ID: id.String()})
	assert.NoError(t, executeErr)

	stub := finalize()
	for _, ev := range stub.Events {
		assert.False(t, startsWith(ev.Name, "cache."),
			"cache.* events must not fire when Cache is nil; got %q", ev.Name)
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Silence unused context import when only some paths use it.
var _ = context.Background
