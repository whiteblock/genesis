package util

import(
	"testing"
)


func TestValidateAscii(t *testing.T) {
	tests := map[string]bool{
		"\u0432\u8977":true,
		"helloworld":false,
		"f\n\r\t\v":false,
		"how are you doing\u8333":true,
	}
	for test,expected := range tests {
		err := ValidateAscii(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateAscii(\"%s\") passed when should have failed",test)
			}else{
				t.Errorf("ValidateAscii(\"%s\") failed when should have passed",test)
			}	
		}
	}
	
}

func TestValidateNormalAscii(t *testing.T) {
	tests := map[string]bool{
		"\u0432\u8977":true,
		"helloworld":false,
		"f\n\r\t\v":true,
		"how are you doing\u8333":true,
	}
	for test,expected := range tests {
		err := ValidateNormalAscii(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateNormalAscii(\"%s\") passed when should have failed",test)
			}else{
				t.Errorf("ValidateNormalAscii(\"%s\") failed when should have passed",test)
			}
			
		}
	}
}

func TestValidateFilePath(t *testing.T) {
	//test --> invalid?
	tests := map[string]bool {
		"../../../":true,
		"genesis.json; rm -rf /":true,
		"config.ini":false,
		"parity/genesis.json":false,
		"\rhello":true,
	}
	for test,expected := range tests {
		err := ValidateFilePath(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateFilePath(\"%s\") passed when should have failed",test)
			}else{
				t.Errorf("ValidateFilePath(\"%s\") failed when should have passed",test)
			}
			
		}
	}
}

func TestValidateCommandLine(t *testing.T) {
	//test --> invalid?
	tests := map[string]bool {
		"../../../":false,
		"genesis.json; rm -rf /":true,
		"config.ini":false,
		"parity/genesis.json":false,
		"\rhello":true,
		"test\";rm -rf /":true,
	}
	for test,expected := range tests {
		err := ValidateCommandLine(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateDockerImage(\"%s\") passed when should have failed",test)
			}else{
				t.Errorf("ValidateDockerImage(\"%s\") failed when should have passed",test)
			}
			
		}
	}
}