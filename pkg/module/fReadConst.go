package module

type fRead_textMarkup_pObj struct {
	begin string
	end   string
}

type fRead_textMarkup_alignObj struct {
	left   fRead_textMarkup_pObj
	right  fRead_textMarkup_pObj
	center fRead_textMarkup_pObj
}
type fRead_textMarkup_styleObj struct {
	throughline fRead_textMarkup_pObj
	underline   fRead_textMarkup_pObj
	italic      fRead_textMarkup_pObj
	bold        fRead_textMarkup_pObj
}

type fRead_textMarkupObj struct {
	ver string

	align fRead_textMarkup_alignObj
	style fRead_textMarkup_styleObj

	del []string
}

var Obj_fRead_textMarkup = fRead_textMarkupObj{
	ver: "1.0",

	align: fRead_textMarkup_alignObj{
		left:   fRead_textMarkup_pObj{"::BL", "::EL"},
		right:  fRead_textMarkup_pObj{"::BR", "::ER"},
		center: fRead_textMarkup_pObj{"::BC", "::EC"},
	},

	style: fRead_textMarkup_styleObj{
		throughline: fRead_textMarkup_pObj{"::BS", "::ES"},
		underline:   fRead_textMarkup_pObj{"::BU", "::EU"},
		italic:      fRead_textMarkup_pObj{"::BI", "::EI"},
		bold:        fRead_textMarkup_pObj{"::BB", "::EB"},
	},
}

func calculate_fRead_allTextMarkup(obj fRead_textMarkupObj) []string {
	var array []string

	array = append(array, obj.align.left.begin)
	array = append(array, obj.align.left.end)

	array = append(array, obj.align.right.begin)
	array = append(array, obj.align.right.end)

	array = append(array, obj.align.center.begin)
	array = append(array, obj.align.center.end)

	//

	array = append(array, obj.style.throughline.begin)
	array = append(array, obj.style.throughline.end)

	array = append(array, obj.style.underline.begin)
	array = append(array, obj.style.underline.end)

	array = append(array, obj.style.italic.begin)
	array = append(array, obj.style.italic.end)

	array = append(array, obj.style.bold.begin)
	array = append(array, obj.style.bold.end)

	return array
}

var Arr_fRead_allTextMarkup []string = calculate_fRead_allTextMarkup(Obj_fRead_textMarkup)
