// gox_textinput.m — GOX TextInput component (UITextField) with full event support

#import <UIKit/UIKit.h>
#import <objc/runtime.h>

// Forward declarations from bridge_core.m
@interface GoxRenderContext : NSObject
@property (nonatomic, strong) NSMutableArray *eventHandlers;
@end

typedef struct {
    const char *tag;
    UIView* (*createView)(NSDictionary *props);
    void (*applyStyle)(UIView *view, NSDictionary *props);
    void (*setContent)(UIView *view, NSString *text, NSDictionary *props);
    void (*wireEvent)(UIView *view, int viewID, GoxRenderContext *ctx);
    void (*updateView)(UIView *view, NSDictionary *oldProps, NSDictionary *newProps);
    void (*childAdded)(UIView *parent, UIView *child, CGRect childFrame);
} GoxComponentDef;

extern UIColor* goxParseColor(NSString *hex);
extern void goxRegisterComponent(GoxComponentDef def);
extern void goxTriggerRerender(void);

// Go exports
extern void GoxHandleTextEvent(int viewID, const char* text);
extern void GoxHandleSubmit(int viewID);
extern void GoxHandleFocus(int viewID);
extern void GoxHandleBlur(int viewID);

// --- TextInput event delegate ---

@interface GoxTextInputDelegate : NSObject <UITextFieldDelegate>
@property (nonatomic, assign) int viewID;
@property (nonatomic, assign) BOOL hasOnChange;
@property (nonatomic, assign) BOOL hasOnSubmit;
@property (nonatomic, assign) BOOL hasOnFocus;
@property (nonatomic, assign) BOOL hasOnBlur;
@end

@implementation GoxTextInputDelegate

- (void)textFieldDidChange:(UITextField *)textField {
    if (!self.hasOnChange) return;
    const char *text = [textField.text UTF8String];
    GoxHandleTextEvent(self.viewID, text ? text : "");
    goxTriggerRerender();
}

- (void)textFieldDidBeginEditing:(UITextField *)textField {
    if (!self.hasOnFocus) return;
    GoxHandleFocus(self.viewID);
}

- (void)textFieldDidEndEditing:(UITextField *)textField {
    if (!self.hasOnBlur) return;
    GoxHandleBlur(self.viewID);
}

- (BOOL)textFieldShouldReturn:(UITextField *)textField {
    if (self.hasOnSubmit) {
        GoxHandleSubmit(self.viewID);
        goxTriggerRerender();
    }
    return YES;
}

- (BOOL)textField:(UITextField *)textField shouldChangeCharactersInRange:(NSRange)range replacementString:(NSString *)string {
    // Enforce maxLength if set
    NSNumber *maxLen = objc_getAssociatedObject(textField, "gox_maxLength");
    if (maxLen) {
        NSUInteger newLength = textField.text.length - range.length + string.length;
        return newLength <= [maxLen unsignedIntegerValue];
    }
    return YES;
}

@end

// --- Helper: apply keyboard type ---
static void applyKeyboardType(UITextField *field, NSString *kbType) {
    if (!kbType) return;
    if ([kbType isEqualToString:@"email"]) field.keyboardType = UIKeyboardTypeEmailAddress;
    else if ([kbType isEqualToString:@"numeric"]) field.keyboardType = UIKeyboardTypeNumberPad;
    else if ([kbType isEqualToString:@"phone"]) field.keyboardType = UIKeyboardTypePhonePad;
    else if ([kbType isEqualToString:@"url"]) field.keyboardType = UIKeyboardTypeURL;
    else if ([kbType isEqualToString:@"decimal"]) field.keyboardType = UIKeyboardTypeDecimalPad;
    else field.keyboardType = UIKeyboardTypeDefault;
}

// --- Helper: apply return key type ---
static void applyReturnKeyType(UITextField *field, NSString *rkType) {
    if (!rkType) return;
    if ([rkType isEqualToString:@"done"]) field.returnKeyType = UIReturnKeyDone;
    else if ([rkType isEqualToString:@"go"]) field.returnKeyType = UIReturnKeyGo;
    else if ([rkType isEqualToString:@"next"]) field.returnKeyType = UIReturnKeyNext;
    else if ([rkType isEqualToString:@"search"]) field.returnKeyType = UIReturnKeySearch;
    else if ([rkType isEqualToString:@"send"]) field.returnKeyType = UIReturnKeySend;
    else field.returnKeyType = UIReturnKeyDefault;
}

// --- Helper: apply auto-capitalize ---
static void applyAutoCapitalize(UITextField *field, NSString *mode) {
    if (!mode) return;
    if ([mode isEqualToString:@"none"]) field.autocapitalizationType = UITextAutocapitalizationTypeNone;
    else if ([mode isEqualToString:@"words"]) field.autocapitalizationType = UITextAutocapitalizationTypeWords;
    else if ([mode isEqualToString:@"all"]) field.autocapitalizationType = UITextAutocapitalizationTypeAllCharacters;
    else field.autocapitalizationType = UITextAutocapitalizationTypeSentences;
}

// --- Component callbacks ---

static UIView* textInputCreate(NSDictionary *props) {
    UITextField *field = [[UITextField alloc] init];
    field.borderStyle = UITextBorderStyleRoundedRect;

    NSString *placeholder = props[@"placeholder"];
    if (placeholder) field.placeholder = placeholder;

    NSString *value = props[@"value"];
    if (value) field.text = value;

    NSNumber *secure = props[@"secure"];
    if (secure && [secure boolValue]) field.secureTextEntry = YES;

    applyKeyboardType(field, props[@"keyboardType"]);
    applyReturnKeyType(field, props[@"returnKeyType"]);
    applyAutoCapitalize(field, props[@"autoCapitalize"]);

    NSNumber *autoCorrect = props[@"autoCorrect"];
    if (autoCorrect) {
        field.autocorrectionType = [autoCorrect boolValue]
            ? UITextAutocorrectionTypeYes
            : UITextAutocorrectionTypeNo;
    }

    NSNumber *editable = props[@"editable"];
    if (editable) field.enabled = [editable boolValue];

    NSNumber *maxLength = props[@"maxLength"];
    if (maxLength) {
        objc_setAssociatedObject(field, "gox_maxLength", maxLength, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
    }

    // Store props for wireEvent to read
    objc_setAssociatedObject(field, "gox_props", props, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    return field;
}

static void textInputApplyStyle(UIView *view, NSDictionary *props) {
    if (![view isKindOfClass:[UITextField class]]) return;
    UITextField *field = (UITextField *)view;

    NSString *color = props[@"color"];
    if (color) {
        UIColor *c = goxParseColor(color);
        if (c) field.textColor = c;
    }

    NSNumber *fontSize = props[@"fontSize"];
    if (fontSize && [fontSize doubleValue] > 0) {
        field.font = [UIFont systemFontOfSize:[fontSize doubleValue]];
    }

    NSString *textAlign = props[@"textAlign"];
    if ([textAlign isEqualToString:@"center"]) field.textAlignment = NSTextAlignmentCenter;
    else if ([textAlign isEqualToString:@"right"]) field.textAlignment = NSTextAlignmentRight;
    else field.textAlignment = NSTextAlignmentLeft;

    NSString *placeholderColor = props[@"placeholderColor"];
    if (placeholderColor && field.placeholder) {
        UIColor *phColor = goxParseColor(placeholderColor);
        if (phColor) {
            field.attributedPlaceholder = [[NSAttributedString alloc]
                initWithString:field.placeholder
                attributes:@{NSForegroundColorAttributeName: phColor}];
        }
    }
}

static void textInputWireEvent(UIView *view, int viewID, GoxRenderContext *ctx) {
    if (![view isKindOfClass:[UITextField class]]) return;
    UITextField *field = (UITextField *)view;

    // Remove old targets to avoid duplicates on re-render
    [field removeTarget:nil action:NULL forControlEvents:UIControlEventEditingChanged];
    field.delegate = nil;

    // Read event flags from the view's associated props
    NSDictionary *props = objc_getAssociatedObject(field, "gox_props");
    BOOL hasOnChange = [props[@"_hasOnChange"] boolValue];
    BOOL hasOnSubmit = [props[@"_hasOnSubmit"] boolValue];
    BOOL hasOnFocus  = [props[@"_hasOnFocus"] boolValue];
    BOOL hasOnBlur   = [props[@"_hasOnBlur"] boolValue];

    if (!hasOnChange && !hasOnSubmit && !hasOnFocus && !hasOnBlur) return;

    GoxTextInputDelegate *delegate = [[GoxTextInputDelegate alloc] init];
    delegate.viewID = viewID;
    delegate.hasOnChange = hasOnChange;
    delegate.hasOnSubmit = hasOnSubmit;
    delegate.hasOnFocus = hasOnFocus;
    delegate.hasOnBlur = hasOnBlur;

    field.delegate = delegate;

    if (hasOnChange) {
        [field addTarget:delegate
                  action:@selector(textFieldDidChange:)
        forControlEvents:UIControlEventEditingChanged];
    }

    // Retain the delegate
    [ctx.eventHandlers addObject:delegate];
}

static void textInputUpdateView(UIView *view, NSDictionary *oldProps, NSDictionary *newProps) {
    if (![view isKindOfClass:[UITextField class]]) return;
    UITextField *field = (UITextField *)view;

    // Store props for wireEvent to read
    objc_setAssociatedObject(field, "gox_props", newProps, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    // Update value — only if it changed and field is not currently being edited
    NSString *newValue = newProps[@"value"];
    NSString *oldValue = oldProps[@"value"];
    if (newValue && ![newValue isEqual:oldValue] && !field.isFirstResponder) {
        field.text = newValue;
    }

    NSString *placeholder = newProps[@"placeholder"];
    if (placeholder) field.placeholder = placeholder;

    NSNumber *secure = newProps[@"secure"];
    if (secure) field.secureTextEntry = [secure boolValue];

    applyKeyboardType(field, newProps[@"keyboardType"]);
    applyReturnKeyType(field, newProps[@"returnKeyType"]);
    applyAutoCapitalize(field, newProps[@"autoCapitalize"]);

    NSNumber *autoCorrect = newProps[@"autoCorrect"];
    if (autoCorrect) {
        field.autocorrectionType = [autoCorrect boolValue]
            ? UITextAutocorrectionTypeYes
            : UITextAutocorrectionTypeNo;
    }

    NSNumber *editable = newProps[@"editable"];
    if (editable) field.enabled = [editable boolValue];

    NSNumber *maxLength = newProps[@"maxLength"];
    if (maxLength) {
        objc_setAssociatedObject(field, "gox_maxLength", maxLength, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
    }

    // Auto-focus on mount
    NSNumber *autoFocus = newProps[@"autoFocus"];
    if (autoFocus && [autoFocus boolValue] && !field.isFirstResponder) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [field becomeFirstResponder];
        });
    }
}

__attribute__((constructor))
static void goxRegisterTextInput(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "TextInput",
        .createView = textInputCreate,
        .applyStyle = textInputApplyStyle,
        .wireEvent = textInputWireEvent,
        .updateView = textInputUpdateView,
    });
}
