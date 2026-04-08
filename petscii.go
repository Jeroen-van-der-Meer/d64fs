package cbm

// PETSCII (PET Standard Code of Information Interchange) is Commodore's
// character encoding. We map each PETSCII byte to a Unicode character. Control
// codes are mapped to U+FFFD.

const rpl  = '\ufffd'  // Replacement char
const nbsp = '\u00a0' // Non-breaking space

var petsciiTable = [256]rune{
	// $00-$1F: Control codes. Printing them will cause a change in screen
	// layout or behavior, not an actual character displayed.
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, '\n', rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,

	// $20-$3F: Space, punctuation, digits. Same as ASCII.
	' ', '!', '"', '#', '$', '%', '&', '\'',
	'(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', ':', ';', '<', '=', '>', '?',

	// $40-$5F: @, uppercase A-Z, and special characters.
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G',
	'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
	'X', 'Y', 'Z', '[', '£', ']', '↑', '←',

	// $60-$7F: Graphic characters. Not used in practice.
	'─', '♠', '🭲', '🭸', '🭷', '🭶', '🭺', '🭱',
	'🭴', '╮', '╰', '╯', '🭼', '╲', '╱', '🭽',
	'🭾', '•', '🭻', '♥', '🭰', '╭', '╳', '○',
	'♣', '🭵', '♦', '┼', '🮌', '│', 'π', '◥',

	// $80-$9F: Control codes.
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,
	rpl, rpl, rpl, rpl, rpl, rpl, rpl, rpl,

	// $A0-$BF: Shifted space + graphic characters.
	nbsp, '▌', '▄', '▔', '▁', '▏', '▒', '▕',
	'🮏', '◤', '🮇', '├', '▗', '└', '┐', '▂',
	'┌', '┴', '┬', '┤', '▎', '▍', '🮈', '🮂',
	'🮃', '▃', '🭿', '▖', '▝', '┘', '▘', '▚',

	// $C0-$DF: More graphic characters.
	'─', '♠', '🭲', '🭸', '🭷', '🭶', '🭺', '🭱',
	'🭴', '╮', '╰', '╯', '🭼', '╲', '╱', '🭽',
	'🭾', '•', '🭻', '♥', '🭰', '╭', '╳', '○',
	'♣', '🭵', '♦', '┼', '🮌', '│', 'π', '◥',

	// $E0-$FF: More graphic characters.
	nbsp, '▌', '▄', '▔', '▁', '▏', '▒', '▕',
	'🮏', '◤', '🮇', '├', '▗', '└', '┐', '▂',
	'┌', '┴', '┬', '┤', '▎', '▍', '🮈', '🮂',
	'🮃', '▃', '🭿', '▖', '▝', '┘', '▘', 'π',
}

// DecodePETSCII converts a PETSCII byte slice to a UTF-8 string.
func DecodePETSCII(b []byte) string {
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = petsciiTable[c]
	}
	return string(runes)
}
