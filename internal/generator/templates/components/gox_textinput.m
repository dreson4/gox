// gox_textinput.m — GOX TextInput component (UITextField)

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

static UIView* textInputCreate(NSDictionary *props) {
    UITextField *field = [[UITextField alloc] init];
    field.borderStyle = UITextBorderStyleRoundedRect;

    NSString *placeholder = props[@"placeholder"];
    if (placeholder) field.placeholder = placeholder;

    NSString *value = props[@"value"];
    if (value) field.text = value;

    NSNumber *secure = props[@"secure"];
    if (secure && [secure boolValue]) field.secureTextEntry = YES;

    NSString *kbType = props[@"keyboardType"];
    if ([kbType isEqualToString:@"email"]) field.keyboardType = UIKeyboardTypeEmailAddress;
    else if ([kbType isEqualToString:@"numeric"]) field.keyboardType = UIKeyboardTypeNumberPad;
    else if ([kbType isEqualToString:@"phone"]) field.keyboardType = UIKeyboardTypePhonePad;
    else if ([kbType isEqualToString:@"url"]) field.keyboardType = UIKeyboardTypeURL;

    return field;
}

__attribute__((constructor))
static void goxRegisterTextInput(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "TextInput",
        .createView = textInputCreate,
    });
}
