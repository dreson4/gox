// gox_scrollview.m — GOX ScrollView component (UIScrollView)

#import <UIKit/UIKit.h>
#import <objc/runtime.h>

@class GoxRenderContext;

typedef struct {
    const char *tag;
    UIView* (*createView)(NSDictionary *props);
    void (*applyStyle)(UIView *view, NSDictionary *props);
    void (*setContent)(UIView *view, NSString *text, NSDictionary *props);
    void (*wireEvent)(UIView *view, int viewID, GoxRenderContext *ctx);
    void (*updateView)(UIView *view, NSDictionary *oldProps, NSDictionary *newProps);
    void (*childAdded)(UIView *parent, UIView *child, CGRect childFrame);
} GoxComponentDef;

extern void goxRegisterComponent(GoxComponentDef def);
extern void goxTriggerRerender(void);

// Go exports
extern void GoxHandleScroll(int viewID, double offset);
extern void GoxHandleScrollEnd(int viewID);

// --- ScrollView delegate ---

@interface GoxScrollViewDelegate : NSObject <UIScrollViewDelegate>
@property (nonatomic, assign) int viewID;
@property (nonatomic, assign) BOOL hasOnScroll;
@property (nonatomic, assign) BOOL hasOnScrollEnd;
@property (nonatomic, assign) BOOL horizontal;
@end

@implementation GoxScrollViewDelegate

- (void)scrollViewDidScroll:(UIScrollView *)scrollView {
    if (!self.hasOnScroll) return;
    double offset = self.horizontal ? scrollView.contentOffset.x : scrollView.contentOffset.y;
    GoxHandleScroll(self.viewID, offset);
    // NO rerender — fires at 60fps, callback is informational only
}

- (void)scrollViewDidEndDecelerating:(UIScrollView *)scrollView {
    if (!self.hasOnScrollEnd) return;
    GoxHandleScrollEnd(self.viewID);
    goxTriggerRerender();
}

- (void)scrollViewDidEndDragging:(UIScrollView *)scrollView willDecelerate:(BOOL)decelerate {
    if (decelerate) return; // will fire in scrollViewDidEndDecelerating
    if (!self.hasOnScrollEnd) return;
    GoxHandleScrollEnd(self.viewID);
    goxTriggerRerender();
}

@end

// --- Helper to apply scroll props ---

static void applyScrollProps(UIScrollView *sv, NSDictionary *props) {
    // scrollEnabled (default: true)
    NSNumber *scrollEnabled = props[@"scrollEnabled"];
    sv.scrollEnabled = scrollEnabled ? [scrollEnabled boolValue] : YES;

    // bounces (default: true)
    NSNumber *bounces = props[@"bounces"];
    sv.bounces = bounces ? [bounces boolValue] : YES;

    // alwaysBounceVertical / alwaysBounceHorizontal
    NSNumber *abv = props[@"alwaysBounceVertical"];
    if (abv) sv.alwaysBounceVertical = [abv boolValue];

    NSNumber *abh = props[@"alwaysBounceHorizontal"];
    if (abh) sv.alwaysBounceHorizontal = [abh boolValue];

    // pagingEnabled
    NSNumber *paging = props[@"pagingEnabled"];
    if (paging) sv.pagingEnabled = [paging boolValue];

    // Scroll indicators (default: true)
    NSNumber *showV = props[@"showsVerticalIndicator"];
    sv.showsVerticalScrollIndicator = showV ? [showV boolValue] : YES;

    NSNumber *showH = props[@"showsHorizontalIndicator"];
    sv.showsHorizontalScrollIndicator = showH ? [showH boolValue] : YES;

    // decelerationRate
    NSString *decel = props[@"decelerationRate"];
    if ([decel isEqualToString:@"fast"]) {
        sv.decelerationRate = UIScrollViewDecelerationRateFast;
    } else {
        sv.decelerationRate = UIScrollViewDecelerationRateNormal;
    }

    // keyboardDismissMode
    NSString *kbMode = props[@"keyboardDismissMode"];
    if ([kbMode isEqualToString:@"onDrag"]) {
        sv.keyboardDismissMode = UIScrollViewKeyboardDismissModeOnDrag;
    } else if ([kbMode isEqualToString:@"interactive"]) {
        sv.keyboardDismissMode = UIScrollViewKeyboardDismissModeInteractive;
    } else {
        sv.keyboardDismissMode = UIScrollViewKeyboardDismissModeNone;
    }

    // Content insets
    CGFloat insetTop = [props[@"contentInsetTop"] doubleValue];
    CGFloat insetBottom = [props[@"contentInsetBottom"] doubleValue];
    if (insetTop > 0 || insetBottom > 0) {
        sv.contentInset = UIEdgeInsetsMake(insetTop, 0, insetBottom, 0);
    }
}

// --- Component functions ---

static UIView* scrollViewCreate(NSDictionary *props) {
    UIScrollView *sv = [[UIScrollView alloc] init];
    sv.clipsToBounds = YES;
    applyScrollProps(sv, props);

    // Store props for wireEvent
    objc_setAssociatedObject(sv, "gox_props", props, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    return sv;
}

static void scrollViewApplyStyle(UIView *view, NSDictionary *props) {
    if (![view isKindOfClass:[UIScrollView class]]) return;
    UIScrollView *sv = (UIScrollView *)view;
    applyScrollProps(sv, props);
}

static void scrollViewWireEvent(UIView *view, int viewID, GoxRenderContext *ctx) {
    if (![view isKindOfClass:[UIScrollView class]]) return;
    UIScrollView *sv = (UIScrollView *)view;

    NSDictionary *props = objc_getAssociatedObject(sv, "gox_props");
    BOOL hasOnScroll = [props[@"_hasOnScroll"] boolValue];
    BOOL hasOnScrollEnd = [props[@"_hasOnScrollEnd"] boolValue];
    BOOL horizontal = [props[@"horizontal"] boolValue];

    if (!hasOnScroll && !hasOnScrollEnd) return;

    GoxScrollViewDelegate *delegate = [[GoxScrollViewDelegate alloc] init];
    delegate.viewID = viewID;
    delegate.hasOnScroll = hasOnScroll;
    delegate.hasOnScrollEnd = hasOnScrollEnd;
    delegate.horizontal = horizontal;

    sv.delegate = delegate;
    // Prevent delegate from being deallocated
    objc_setAssociatedObject(sv, "gox_scrollDelegate", delegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}

static void scrollViewUpdateView(UIView *view, NSDictionary *oldProps, NSDictionary *newProps) {
    if (![view isKindOfClass:[UIScrollView class]]) return;
    UIScrollView *sv = (UIScrollView *)view;
    applyScrollProps(sv, newProps);

    // Store updated props for wireEvent
    objc_setAssociatedObject(sv, "gox_props", newProps, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}

static void scrollViewChildAdded(UIView *parent, UIView *child, CGRect childFrame) {
    if (![parent isKindOfClass:[UIScrollView class]]) return;
    UIScrollView *sv = (UIScrollView *)parent;
    CGFloat maxY = CGRectGetMaxY(childFrame);
    CGFloat maxX = CGRectGetMaxX(childFrame);
    if (maxY > sv.contentSize.height || maxX > sv.contentSize.width) {
        sv.contentSize = CGSizeMake(
            MAX(sv.contentSize.width, maxX),
            MAX(sv.contentSize.height, maxY)
        );
    }
}

__attribute__((constructor))
static void goxRegisterScrollView(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "ScrollView",
        .createView = scrollViewCreate,
        .applyStyle = scrollViewApplyStyle,
        .wireEvent = scrollViewWireEvent,
        .updateView = scrollViewUpdateView,
        .childAdded = scrollViewChildAdded,
    });
}
