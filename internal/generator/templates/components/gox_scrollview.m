// gox_scrollview.m — GOX ScrollView component (UIScrollView)

#import <UIKit/UIKit.h>

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

static UIView* scrollViewCreate(NSDictionary *props) {
    UIScrollView *scrollView = [[UIScrollView alloc] init];
    return scrollView;
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
        .childAdded = scrollViewChildAdded,
    });
}
